package pac

type PACRule struct {
	Name        string `json:"name"`
	Proxy       string `json:"proxy"`
	SOCKS4      string `json:"socks4"`
	SOCKS5      string `json:"socks5"`
	LocalRules  string `json:"local_rule_file"`
	RemoteRules string `json:"remote_rule_file"`
}

type PAC struct {
	Rules   []PACRule `json:"rules"`
	Address string    `json:"address"`
}
