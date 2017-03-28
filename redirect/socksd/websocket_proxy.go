package socksd

import (
	"github.com/ssoor/youniverse/log"
	"golang.org/x/net/websocket"
)

func WebsocketEcho(wsServer *websocket.Conn) {
	var err error
	var reply []byte
	var wsClient *websocket.Conn
	var strWebsocketAddr = wsServer.Request().URL.String()

	if wsClient, err = websocket.Dial(strWebsocketAddr, "", wsServer.Request().Header.Get("Origin")); nil != err {
		log.Error("Websocket connection", strWebsocketAddr+"failed, err:", err)
		return
	}

	go func() { //
		var wsBuff []byte

		for {
			if err := websocket.Message.Receive(wsClient, &wsBuff); nil != err {
				break
			}

			if err := websocket.Message.Send(wsServer, wsBuff); nil != err {
				break
			}
		}
	}()

	for {
		if err := websocket.Message.Receive(wsServer, &reply); nil != err {

			break
		}

		if err := websocket.Message.Send(wsClient, reply); nil != err {
			break
		}
	}
}
