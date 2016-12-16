package pac

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ssoor/youniverse/log"
)

var pacTemplate = `
var hasOwnProperty = Object.hasOwnProperty;
{{range .Rules}}
var proxy_{{.Name}} = "{{if .Proxy}}PROXY {{.Proxy}};{{end}}{{if .SOCKS5}}SOCKS5 {{.SOCKS5}};{{end}}{{if .SOCKS4}}SOCKS4 {{.SOCKS4}};{{end}}DIRECT;";

var domains_{{.Name}} = {
  {{.LocalRules}}
};
{{end}}

function FindProxyForURL(url, host) {
	if (url.substring(0, 6) == "https:") {
        return 'DIRECT';
    }
	if (isPlainHostName(host) || host === '127.0.0.1' || host === 'localhost') {
        return 'DIRECT';
    }

    var suffix;
    var pos = host.lastIndexOf('.');
    while(1) {
        suffix = host.substring(pos + 1);
{{range .Rules}}
	    if (hasOwnProperty.call(domains_{{.Name}}, suffix)) {
	        return proxy_{{.Name}};
	    }
{{end}}
        if (pos <= 0) {
            break;
        }

        pos = host.lastIndexOf('.', pos - 1);
    }
    return 'DIRECT';
}
`

type PACGenerator struct {
	filter map[string]int
	Rules  []PACRule
}

func NewPACGenerator(pacRules []PACRule) *PACGenerator {
	rules := make([]PACRule, len(pacRules))

	copy(rules, pacRules)

	return &PACGenerator{
		Rules:  rules,
		filter: make(map[string]int),
	}
}

func (p *PACGenerator) Generate(index int, rules []string) ([]byte, error) {

	for _, v := range rules {
		if _, ok := p.filter[v]; !ok {
			p.filter[v] = index
		}
	}

	data := struct {
		Rules []PACRule
	}{
		Rules: p.Rules,
	}

	for i := 0; i < len(data.Rules); i++ {
		data.Rules[i].LocalRules = ""
	}

	for host, index := range p.filter {
		data.Rules[index].LocalRules += fmt.Sprintf(",'%s' : 1", host)
	}

	for i := 0; i < len(data.Rules); i++ {
		strlen := len(data.Rules[i].LocalRules)
		if strlen > 0 {
			data.Rules[i].LocalRules = data.Rules[i].LocalRules[1:strlen]
		}
	}

	t, err := template.New("proxy.pac").Parse(pacTemplate)

	if err != nil {
		log.Error("failed to parse pacTempalte, err:", err)
	}
	buff := bytes.NewBuffer(nil)
	err = t.Execute(buff, &data)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return buff.Bytes(), nil
}
