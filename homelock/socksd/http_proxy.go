package socksd

import (
	"net"
	"net/http"

	"github.com/ssoor/socks"
)

func StartHTTPProxy(conf Proxy, router socks.Dialer, data []byte) {
	if conf.HTTP != "" {
		listener, err := net.Listen("tcp", conf.HTTP)
		if err != nil {
			ErrLog.Println("net.Listen at ", conf.HTTP, " failed, err:", err)
			return
		}

		defer listener.Close()
		httpProxy := socks.NewHTTPProxy(router, data)
		http.Serve(listener, httpProxy)
	}
}

func StartEncodeHTTPProxy(conf Proxy, router socks.Dialer, data []byte) {
	if conf.HTTP != "" {
		listener, err := net.Listen("tcp", conf.HTTP)
		if err != nil {
			ErrLog.Println("net.Listen at ", conf.HTTP, " failed, err:", err)
			return
		}

		listener = socks.NewHTTPEncodeListener(listener)

		defer listener.Close()
		httpProxy := socks.NewHTTPProxy(router, data)
		http.Serve(listener, httpProxy)
	}
}