package socksd

type Upstream struct {
	Type     string `json:"type"`
	Crypto   string `json:"crypto"`
	Password string `json:"password"`
	Address  string `json:"address"`
}
