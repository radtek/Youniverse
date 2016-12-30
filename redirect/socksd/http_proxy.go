package socksd

import (
	"net"
	"net/http"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
)

func StartHTTPProxy(addr string, router socks.Dialer, data []byte) {
	httpProxy := socks.NewHTTPProxy("http", router, NewHTTPTransport(router, data))
	if err := http.ListenAndServe(addr, httpProxy); nil != err {
		log.Error("Start HTTP proxy at ", addr, " failed, err:", err)
	}
}

func StartEncodeHTTPProxy(addr string, router socks.Dialer, data []byte) {
	if addr != "" {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Error("net.Listen at ", addr, " failed, err:", err)
			return
		}

		listener = NewHTTPEncodeListener(listener)

		defer listener.Close()
		httpProxy := socks.NewHTTPProxy("http", router, NewHTTPTransport(router, data))
		http.Serve(listener, httpProxy)
	}
}
