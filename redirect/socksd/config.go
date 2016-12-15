package socksd

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ssoor/youniverse/redirect/pac"
)

type Upstream struct {
	Type     string `json:"type"`
	Crypto   string `json:"crypto"`
	Password string `json:"password"`
	Address  string `json:"address"`
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
	PAC     pac.PAC `json:"pac"`
	Proxies []Proxy `json:"proxies"`
}

func LoadConfig(s string) (*Config, error) {
	data, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	cfgGroup := &Config{}
	if err = json.Unmarshal(data, cfgGroup); err != nil {
		return nil, err
	}
	return cfgGroup, nil
}
