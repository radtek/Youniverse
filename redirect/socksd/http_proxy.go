package socksd

import (
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
)

type HTTPHandler struct {
	proxy     *socks.HTTPProxy
	websocket websocket.Handler
}

func (h *HTTPHandler) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	request.URL.Scheme = "http"
	request.URL.Host = request.Host

	var handler http.Handler = h.proxy

	if strings.EqualFold(request.Header.Get("Connection"), "Upgrade") && strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
		handler = h.websocket
		request.URL.Scheme = "ws"
	}

	handler.ServeHTTP(rw, request)
}

func StartHTTPProxy(addr string, router socks.Dialer, tran *HTTPTransport) {
	handler := &HTTPHandler{
		websocket: websocket.Handler(WebsocketEcho),
		proxy:     socks.NewHTTPProxy("http", router, tran),
	}

	if err := http.ListenAndServe(addr, handler); nil != err {
		log.Error("Start HTTP proxy at ", addr, " failed, err:", err)
	}
}

func StartEncodeHTTPProxy(addr string, router socks.Dialer, tran *HTTPTransport) {
	if addr != "" {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Error("net.Listen at ", addr, " failed, err:", err)
			return
		}

		listener = NewHTTPEncodeListener(listener)
		defer listener.Close()

		handler := &HTTPHandler{
			websocket: websocket.Handler(WebsocketEcho),
			proxy:     socks.NewHTTPProxy("http", router, tran),
		}

		http.Serve(listener, handler)
	}
}
