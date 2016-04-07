package homelock

import (
	"errors"
	"net"
	"strconv"

	. "github.com/ssoor/youniverse/homelock/socksd"
)

var (
	ErrorSocketUnavailable error = errors.New("socket port not find")
)

func SocketSelectPort(port_type string, port_base int) (int, error) {

	for ; port_base < 65536; port_base++ {

		tcpListener, err := net.Listen(port_type, ":"+strconv.Itoa(port_base))

		if err == nil {
			tcpListener.Close()
			return port_base, nil
		}
	}

	return 0, ErrorSocketUnavailable
}

func CreateSocksdProxy(userGUID string, upstream []Upstream) (Proxy, error) {
	portHttp, _ := SocketSelectPort("tcp", 60000)
	portSocket4, _ := SocketSelectPort("tcp", portHttp+1)
	portSocket5, _ := SocketSelectPort("tcp", portSocket4+1)

	if 0 == portHttp || 0 == portSocket5 {
		return Proxy{}, ErrorSocketUnavailable
	}

	proxy := Proxy{
		HTTP:      ":" + strconv.Itoa(portHttp),
		SOCKS4:    ":" + strconv.Itoa(portSocket4),
		SOCKS5:    ":" + strconv.Itoa(portSocket5),
		Upstreams: upstream,
	}

	return proxy, nil
}

func CreateSocksdPAC(guid string, addr string,proxie Proxy, upstream Upstream,bricksURL string) (*PAC, error) {
	portHttp, _ := SocketSelectPort("tcp", 60000)
	portSocket5, _ := SocketSelectPort("tcp", portHttp+1)

	if 0 == portHttp || 0 == portSocket5 {
		return nil, ErrorSocketUnavailable
	}

	pac := &PAC{
		Address:  addr,
		Upstream: upstream,
		Rules: []PACRule{
			{
				Name:   "default_proxy",
				Proxy:  proxie.HTTP,
				SOCKS5: proxie.SOCKS5,
				//LocalRules: "default_rules.txt",
				RemoteRules: bricksURL,
			},
		},
	}

	return pac, nil
}

