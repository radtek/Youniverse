package redirect

import (
	"encoding/binary"
	"net"

	"github.com/ssoor/youniverse/assistant"
)

func SocketCreateSockAddr(host string, port uint16) (addrSocket assistant.SOCKADDR_IN) {
	binary.BigEndian.PutUint16(addrSocket.Sin_port[0:], port)

	ipv4 := net.ParseIP(host)

	ipv4 = ipv4.To4()
	buff := make([]byte, 0)

	buff = append(buff, ipv4...)

	copy(addrSocket.Sin_addr[0:], buff)

	return addrSocket
}
