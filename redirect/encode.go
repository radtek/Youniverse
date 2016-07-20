package redirect

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

var (
	// Library
	libhttpredirect syscall.Handle

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

func LoadDLL() error {
	// Library
	libhttpredirect, err := syscall.LoadLibrary("youniverse.dll")
	if err != nil {
		return errors.New(fmt.Sprint("Load libaray filed:", err))
	}

	// Functions
	startBusiness, err = syscall.GetProcAddress(libhttpredirect, "StartBusiness")
	if err != nil {
		return errors.New(fmt.Sprint("Query function StartBusiness filed:", err))
	}

	setBusinessData, err = syscall.GetProcAddress(libhttpredirect, "SetBusinessData")
	if err != nil {
		return errors.New(fmt.Sprint("Query function SetBusinessData filed:", err))
	}

	return nil
}

func UnoadDLL() {
	startBusiness = 0
	setBusinessData = 0
	syscall.FreeLibrary(syscall.Handle(libhttpredirect))
}

func SocketCreateSockAddr(host string, port uint16) (addrSocket SOCKADDR_IN) {
	binary.BigEndian.PutUint16(addrSocket.sin_port[0:], port)

	ipv4 := net.ParseIP(host)

	ipv4 = ipv4.To4()
	buff := make([]byte, 0)

	buff = append(buff, ipv4...)

	copy(addrSocket.sin_addr[0:], buff)

	return addrSocket
}

func StartBusiness() int32 {
	ret, _, _ := syscall.Syscall(startBusiness, 1,
		uintptr(unsafe.Pointer(nil)),
		0, 0)

	return int32(ret)
}

func SetBusinessData(addrPACSocket SOCKADDR_IN, addrEncodeSocket SOCKADDR_IN) int32 {
	ret, _, _ := syscall.Syscall(setBusinessData, 2,
		uintptr(unsafe.Pointer(&addrPACSocket)),
		uintptr(unsafe.Pointer(&addrEncodeSocket)),
		0)

	return int32(ret)
}
