package internest

type Settings struct {
	APIPort    int    `json:"api_port"`
	SignURL    string `json:"sign_url"`
	EnforceURL string `json:"enforce_url"`
}
