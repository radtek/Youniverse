package socksd

import (
	"crypto/tls"
	"net/http"
	"time"

	"strings"

	"golang.org/x/net/websocket"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
)

type HTTPSProxyHandler struct {
	proxy     *socks.HTTPProxy
	websocket websocket.Handler
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

func (h *HTTPSProxyHandler) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	request.URL.Scheme = "https"
	request.URL.Host = request.Host

	var handler http.Handler = h.proxy

	if strings.EqualFold(request.Header.Get("Connection"), "Upgrade") && strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
		handler = h.websocket
		request.URL.Scheme = "wss"
	}

	handler.ServeHTTP(rw, request)
}

func HTTPSGetCertificate(clientHello *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
	if cert, err = QueryTlsCertificate(clientHello.ServerName); nil == err {
		return cert, err
	}

	return CreateTlsCertificate(nil, clientHello.ServerName, -(365 * 24 * time.Hour), 200)
}

func StartHTTPSProxy(addr string, router socks.Dialer, tran *HTTPTransport) {
	serverHTTPS := &http.Server{
		ErrorLog: log.Warn,
		TLSConfig: &tls.Config{
			GetCertificate: HTTPSGetCertificate,
		},

		Addr: addr,
		Handler: &HTTPSProxyHandler{
			websocket: websocket.Handler(WebsocketEcho),
			proxy:     socks.NewHTTPProxy("https", router, tran),
		},
	}

	if err := serverHTTPS.ListenAndServeTLS("", ""); nil != err {
		log.Error("Start HTTP proxy at ", addr, " failed, err:", err)
	}
}
