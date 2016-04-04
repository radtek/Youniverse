package homelock

import (
	"encoding/binary"
	"net"
	"syscall"
	"unsafe"

	"github.com/ssoor/winapi"
)

var (
	// Library
	libhttpredirect uintptr

	// Functions
	startBusiness   uintptr
	setBusinessData uintptr
)

type SOCKADDR_IN struct {
	sin_family int16
	sin_port   [2]byte
	sin_addr   [4]byte
	sin_zero   [8]byte
}

func LoadDLL() {
	// Library
	libhttpredirect = winapi.MustLoadLibrary("httpredirect.dll")

	// Functions
	startBusiness = winapi.MustGetProcAddress(libhttpredirect, "StartBusiness")
	setBusinessData = winapi.MustGetProcAddress(libhttpredirect, "SetBusinessData")
}

func SocketCreateSockAddr(host string, port uint16) (addrSocket SOCKADDR_IN) {
	binary.BigEndian.PutUint16(addrSocket.sin_port[0:], port)

	ipv4 := net.ParseIP("127.0.0.1")

	ipv4 = ipv4.To4()
	buff := make([]byte, 0)

	buff = append(buff, ipv4...)

	copy(addrSocket.sin_addr[0:], buff)

	return addrSocket
}

func StartBusiness() int32 {
	ret, _, _ := syscall.Syscall6(startBusiness, 1,
		uintptr(unsafe.Pointer(nil)),
		0, 0, 0, 0, 0)

	return int32(ret)
}

func SetBusinessData(addrPACSocket SOCKADDR_IN, addrEncodeSocket SOCKADDR_IN) int32 {
	ret, _, _ := syscall.Syscall6(setBusinessData, 2,
		uintptr(unsafe.Pointer(&addrPACSocket)),
		uintptr(unsafe.Pointer(&addrEncodeSocket)),
		0, 0, 0, 0)

	return int32(ret)
}
