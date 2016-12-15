package socksd

import (
	"net"
	"net/http"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
)

func StartHTTPProxy(conf Proxy, router socks.Dialer, data []byte) {
	httpProxy := socks.NewHTTPProxy("http", router, NewHTTPTransport(router, data))
	if err := http.ListenAndServe(conf.HTTP, httpProxy); nil != err {
		log.Error("Start HTTP proxy at ", conf.HTTP, " failed, err:", err)
	}
}

func StartEncodeHTTPProxy(conf Proxy, router socks.Dialer, data []byte) {
	if conf.HTTP != "" {
		listener, err := net.Listen("tcp", conf.HTTP)
		if err != nil {
			log.Error("net.Listen at ", conf.HTTP, " failed, err:", err)
			return
		}

		listener = NewHTTPEncodeListener(listener)

		defer listener.Close()
		httpProxy := socks.NewHTTPProxy("http", router, NewHTTPTransport(router, data))
		http.Serve(listener, httpProxy)
	}
}
