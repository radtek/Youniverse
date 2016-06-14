package socksd

import (
	"net"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
)

type DecorateClient struct {
	forward    socks.Dialer
	decorators []ConnDecorator
}

func NewDecorateClient(forward socks.Dialer, ds ...ConnDecorator) *DecorateClient {
	d := &DecorateClient{
		forward: forward,
	}
	d.decorators = append(d.decorators, ds...)
	return d
}

func (d *DecorateClient) Dial(network, address string) (net.Conn, error) {
	conn, err := d.forward.Dial(network, address)
	if err != nil {
		log.Error("DecorateClient forward.Dial failed, err:", err, address)
		return nil, err
	}
	dconn, err := DecorateConn(conn, d.decorators...)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return dconn, nil
}
