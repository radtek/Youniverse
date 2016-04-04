package homelock

import (
	"errors"
	"net"
	"strconv"
)
type Upstream struct {
	Type     string `json:"type"`
	Crypto   string `json:"crypto"`
	Password string `json:"password"`
	Address  string `json:"address"`
}

type PACRule struct {
	Name        string `json:"name"`
	Proxy       string `json:"proxy"`
	SOCKS4      string `json:"socks4"`
	SOCKS5      string `json:"socks5"`
	LocalRules  string `json:"local_rule_file"`
	RemoteRules string `json:"remote_rule_file"`
}

type PAC struct {
	Rules    []PACRule `json:"rules"`
	Address  string    `json:"address"`
	Upstream Upstream  `json:"upstream"`
}

type Proxy struct {
	HTTP            string     `json:"http"`
	SOCKS4          string     `json:"socks4"`
	SOCKS5          string     `json:"socks5"`
	Crypto          string     `json:"crypto"`
	Password        string     `json:"password"`
	DNSCacheTimeout int        `json:"dnsCacheTimeout"`
	Upstreams       []Upstream `json:"upstreams"`
}

type Config struct {
	PAC     PAC     `json:"pac"`
	Proxies []Proxy `json:"proxies"`
}

var (
	ErrorSocketUnavailable error = errors.New("socket port not find")
)

func SocketSelectPort(port_type string, port_base int) (int,error) {

	for ; port_base < 65536; port_base++ {

		tcpListener, err := net.Listen(port_type, ":"+strconv.Itoa(port_base))

		if err == nil {
			tcpListener.Close()
			return port_base,nil
		}
	}
    
	return 0, ErrorSocketUnavailable
}

func createSocksPACRule(userGUID string) (*PACRule, error) {
	portHttp,_ := SocketSelectPort("tcp", 60000)
	portSocket5,_ := SocketSelectPort("tcp", portHttp+1)

	if 0 == portHttp || 0 == portSocket5 {
		return nil, ErrorSocketUnavailable
	}

	rule := &PACRule{
		Name:   "default_proxy",
		Proxy:  "127.0.0.1:" + strconv.Itoa(portHttp),
		SOCKS5: "127.0.0.1:" + strconv.Itoa(portSocket5),
		//LocalRules: "default_rules.txt",
		RemoteRules: "http://120.26.80.61/issued/bricks/20160308/" + userGUID + ".bricks",
	}

	return rule, nil
}

var (
	ErrorUpstreamUnknown error = errors.New("Upstream is nil")
)

func CreateSocksConfig(pacAddress string, upstreams []Upstream, userGUID string) (*Config, error) {

	if len(upstreams) < 1 {
		return nil, ErrorUpstreamUnknown
	}

	var config Config

	config.Proxies = []Proxy{}

	config.PAC.Rules = []PACRule{}
	config.PAC.Address = pacAddress
	//config.PAC.Upstream = upstreams[0]

	rule, err := createSocksPACRule(userGUID)

	if err != nil {
		return nil, err
	}

	proxie := Proxy{
		HTTP:      rule.Proxy,
		SOCKS5:    rule.SOCKS5,
		Upstreams: upstreams,
	}

	config.Proxies = append(config.Proxies, proxie)
	config.PAC.Rules = append(config.PAC.Rules, *rule)

	return &config, nil
}
