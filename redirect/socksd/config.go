package socksd

type Upstream struct {
	Type     string `json:"type"`
	Crypto   string `json:"crypto"`
	Password string `json:"password"`
	Address  string `json:"address"`
}

type Setting struct {
	IntervalTime int        `json:"interval_time"`
	DNSCacheTime int        `json:"dnscache_time"`
	Upstreams    []Upstream `json:"services"`
}
