package socksd

import (
	"crypto/tls"
	"net/http"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
)

type HTTPSProxyHandler struct {
	proxy *socks.HTTPProxy
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; http://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (h *HTTPSProxyHandler) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	h.proxy.ServeHTTP(rw, request)
}

func HTTPSGetCertificate(clientHello *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
	if cert, err = QueryTlsCertificate(clientHello.ServerName); nil == err {
		return cert, err
	}

	return CreateTlsCertificate(clientHello.ServerName)
}

func StartHTTPSProxy(conf Proxies, router socks.Dialer, data []byte) {
	serverHTTPS := &http.Server{
		Addr:    conf.HTTPS,
		Handler: &HTTPSProxyHandler{proxy: socks.NewHTTPProxy("https", router, NewHTTPTransport(router, data))},
		TLSConfig: &tls.Config{
			GetCertificate: HTTPSGetCertificate,
		},
	}

	if err := serverHTTPS.ListenAndServeTLS("", ""); nil != err {
		log.Error("Start HTTP proxy at ", conf.HTTP, " failed, err:", err)
	}
}