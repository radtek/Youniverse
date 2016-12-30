package socksd

import (
	"errors"
	"strings"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
)

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

func BuildUpstreamRouter(timeoutDNSCache int, upstreams []Upstream) socks.Dialer {
	var allForward []socks.Dialer
	for _, upstream := range upstreams {
		var forward socks.Dialer
		var err error
		forward = NewDecorateDirect(timeoutDNSCache)
		forward, err = BuildUpstream(upstream, forward)
		if err != nil {
			log.Error("failed to BuildUpstream, err:", err)
			continue
		}
		allForward = append(allForward, forward)
	}
	if len(allForward) == 0 {
		router := NewDecorateDirect(timeoutDNSCache)
		allForward = append(allForward, router)
	}
	return NewUpstreamDialer(allForward)
}
