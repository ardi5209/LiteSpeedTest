package main

import (
	"context"
	"errors"
	"regexp"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/component/resolver"
	_ "github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/dns"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/proxy"
	"github.com/xxf098/lite-proxy/tunnel"
	"github.com/xxf098/lite-proxy/tunnel/adapter"
	"github.com/xxf098/lite-proxy/tunnel/http"
	"github.com/xxf098/lite-proxy/tunnel/socks"
)

func startInstance(c Config) (*proxy.Proxy, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "LocalHost", c.LocalHost)
	ctx = context.WithValue(ctx, "LocalPort", c.LocalPort)
	adapterServer, err := adapter.NewServer(ctx, nil)
	if err != nil {
		return nil, err
	}
	httpServer, err := http.NewServer(ctx, adapterServer)
	if err != nil {
		return nil, err
	}
	socksServer, err := socks.NewServer(ctx, adapterServer)
	if err != nil {
		return nil, err
	}
	sources := []tunnel.Server{httpServer, socksServer}
	sink, err := createSink(ctx, c.Link)
	if err != nil {
		return nil, err
	}
	setDefaultResolver()
	p := proxy.NewProxy(ctx, cancel, sources, sink)
	return p, nil
}

func createSink(ctx context.Context, link string) (tunnel.Client, error) {
	var d outbound.Dialer
	r := regexp.MustCompile("(?i)^(vmess|trojan|ss|ssr)://.+")
	matches := r.FindStringSubmatch(link)
	if len(matches) < 2 {
		return nil, common.NewError("Not Suported Link")
	}
	creator, err := outbound.GetDialerCreator(matches[1])
	if err != nil {
		return nil, err
	}
	d, err = creator(link)
	if err != nil {
		return nil, err
	}
	if d != nil {
		return proxy.NewClient(ctx, d), nil
	}

	return nil, errors.New("not supported link")
}

func setDefaultResolver() {
	resolver.DefaultResolver = dns.DefaultResolver()
}
