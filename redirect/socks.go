package redirect

import (
	"errors"
	"strconv"

	"github.com/ssoor/youniverse/common"
	"github.com/ssoor/youniverse/redirect/pac"
	. "github.com/ssoor/youniverse/redirect/socksd"
)

var (
	ErrorSocketUnavailable error = errors.New("socket port not find")
)

func CreateSocksdProxy(userGUID string, ipAddr string, upstream []Upstream) (Proxy, error) {
	portHttp, _ := common.SocketSelectPort("tcp")
	portSocket4, _ := common.SocketSelectPort("tcp")
	portSocket5, _ := common.SocketSelectPort("tcp")

	if 0 == portHttp || 0 == portSocket5 {
		return Proxy{}, ErrorSocketUnavailable
	}

	proxy := Proxy{
		HTTP:      ipAddr + ":" + strconv.Itoa(portHttp),
		SOCKS4:    ipAddr + ":" + strconv.Itoa(portSocket4),
		SOCKS5:    ipAddr + ":" + strconv.Itoa(portSocket5),
		Upstreams: upstream,
	}

	return proxy, nil
}

func CreateSocksdPAC(guid string, addr string, proxie Proxy, upstream Upstream, bricksURL string) (*pac.PAC, error) {
	portHttp, _ := common.SocketSelectPort("tcp")
	portSocket5, _ := common.SocketSelectPort("tcp")

	if 0 == portHttp || 0 == portSocket5 {
		return nil, ErrorSocketUnavailable
	}

	cfgPAC := &pac.PAC{
		Address: addr,
		Rules: []pac.PACRule{
			{
				Name:   "default_proxy",
				Proxy:  proxie.HTTP,
				SOCKS5: proxie.SOCKS5,
				//LocalRules: "default_rules.txt",
				RemoteRules: bricksURL,
			},
		},
	}

	return cfgPAC, nil
}
