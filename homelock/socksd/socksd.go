package socksd

import (
	"errors"
	"net"
	"strings"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/api"
)

func StartSocksd(guid string, encode bool, conf *Config) {
	url := "http://120.26.80.61/issued/rules/20160308/" + guid + ".rules"

	srules, err := api.GetURL(url)
	if err != nil {
		InfoLog.Printf("Load srules: %s failed, err: %s\n", url, err)
		return
	}
	InfoLog.Printf("Load srules: %s succeeded\n", url)

	for _, c := range conf.Proxies {
		router := BuildUpstreamRouter(c)

		if encode {
			go StartEncodeHTTPProxy(c, router, []byte(srules))
		} else {
			go StartHTTPProxy(c, router, []byte(srules))
		}

		go runSOCKS4Server(c, router)
		go runSOCKS5Server(c, router)
	}

	StartPACServer(conf.PAC)
}

func BuildUpstream(upstream Upstream, forward socks.Dialer) (socks.Dialer, error) {
	cipherDecorator := NewCipherConnDecorator(upstream.Crypto, upstream.Password)
	forward = NewDecorateClient(forward, cipherDecorator)

	switch strings.ToLower(upstream.Type) {
	case "socks5":
		{
			return socks.NewSocks5Client("tcp", upstream.Address, "", "", forward)
		}
	case "shadowsocks":
		{
			return socks.NewShadowSocksClient("tcp", upstream.Address, forward)
		}
	}
	return nil, errors.New("unknown upstream type" + upstream.Type)
}

func BuildUpstreamRouter(conf Proxy) socks.Dialer {
	var allForward []socks.Dialer
	for _, upstream := range conf.Upstreams {
		var forward socks.Dialer
		var err error
		forward = NewDecorateDirect(conf.DNSCacheTimeout)
		forward, err = BuildUpstream(upstream, forward)
		if err != nil {
			ErrLog.Println("failed to BuildUpstream, err:", err)
			continue
		}
		allForward = append(allForward, forward)
	}
	if len(allForward) == 0 {
		router := NewDecorateDirect(conf.DNSCacheTimeout)
		allForward = append(allForward, router)
	}
	return NewUpstreamDialer(allForward)
}

func runSOCKS4Server(conf Proxy, forward socks.Dialer) {
	if conf.SOCKS4 != "" {
		listener, err := net.Listen("tcp", conf.SOCKS4)
		if err != nil {
			ErrLog.Println("net.Listen failed, err:", err, conf.SOCKS4)
			return
		}
		cipherDecorator := NewCipherConnDecorator(conf.Crypto, conf.Password)
		listener = NewDecorateListener(listener, cipherDecorator)
		socks4Svr, err := socks.NewSocks4Server(forward)
		if err != nil {
			listener.Close()
			ErrLog.Println("socks.NewSocks4Server failed, err:", err)
		}

		defer listener.Close()
		socks4Svr.Serve(listener)
	}
}

func runSOCKS5Server(conf Proxy, forward socks.Dialer) {
	if conf.SOCKS5 != "" {
		listener, err := net.Listen("tcp", conf.SOCKS5)
		if err != nil {
			ErrLog.Println("net.Listen failed, err:", err, conf.SOCKS5)
			return
		}
		cipherDecorator := NewCipherConnDecorator(conf.Crypto, conf.Password)
		listener = NewDecorateListener(listener, cipherDecorator)
		socks5Svr, err := socks.NewSocks5Server(forward)
		if err != nil {
			listener.Close()
			ErrLog.Println("socks.NewSocks5Server failed, err:", err)
			return
		}

		defer listener.Close()
		socks5Svr.Serve(listener)
	}
}
