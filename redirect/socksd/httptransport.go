package socksd

import (
	"net"
	"net/http"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
)

type HTTPTransport struct {
	http.Transport

	Rules *SRules
}

func NewHTTPTransport(forward socks.Dialer, jsondata []byte) *HTTPTransport {

	transport := &HTTPTransport{
		Rules: NewSRules(forward),
		Transport: http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return forward.Dial(network, addr)
			},
		},
	}

	if err := transport.Rules.ResolveJson(jsondata); nil != err {
		log.Error("Transport resolve json rule failed, err:", err)
	}

	return transport
}

func (this *HTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	tranpoort, resp := this.Rules.ResolveRequest(req)

	if nil != resp {
		return resp, nil
	}

	req.Header.Del("X-Forwarded-For")
	resp, err = tranpoort.RoundTrip(req)

	if err != nil {
		log.Error("tranpoort round trip err:", err)
		return
	}

	resp = this.Rules.ResolveResponse(req, resp)

	return
}
