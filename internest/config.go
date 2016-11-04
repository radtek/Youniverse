package internest

type jsonURLNested struct {
	Type      int    `json:"type"`
	URL       string `json:"url"`
	Path      string `json:"path"`
	Title     string `json:"title"`
	ScriptURL string `json:"script_url"`
}

type jsonHtmlNested struct {
	Data   string            `json:"data"`
	Status int               `json:"status"`
	Path   string            `json:"path"`
	Header map[string]string `json:"header"`
}

type Settings struct {
	APIPort    int              `json:"api_port"`
	URLNested  []jsonURLNested  `json:"url_nested"`
	HtmlNested []jsonHtmlNested `json:"html_nested"`
}
