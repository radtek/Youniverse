package assistant

import (
	"syscall"
	"unsafe"
)

type SOCKADDR_IN struct {
	Sin_family int16
	Sin_port   [2]byte
	Sin_addr   [4]byte
	Sin_zero   [8]byte
}

func StartBusiness() (int32, error) {
	libhttpredirect, err := syscall.LoadLibrary("youniverse.dll")
	if err != nil {
		return 0, err
	}

	addrFuncation, err := syscall.GetProcAddress(libhttpredirect, "StartBusiness")
	if err != nil {
		return 0, err
	}

	ret, _, _ := syscall.Syscall(addrFuncation, 1,
		uintptr(unsafe.Pointer(nil)),
		0, 0)

	syscall.FreeLibrary(syscall.Handle(libhttpredirect))

	return int32(ret), nil
}

func SetAPIPort(port int) (int32, error) {
	libhttpredirect, err := syscall.LoadLibrary("youniverse.dll")
	if err != nil {
		return 0, err
	}

	addrFuncation, err := syscall.GetProcAddress(libhttpredirect, "SetAPIPort")
	if err != nil {
		return 0, err
	}

	ret, _, _ := syscall.Syscall(addrFuncation, 1,
		uintptr(unsafe.Pointer(&port)),
		0, 0)

	syscall.FreeLibrary(syscall.Handle(libhttpredirect))

	return int32(ret), nil
}

func SetBusinessData(addrPACSocket SOCKADDR_IN, addrEncodeSocket SOCKADDR_IN) (int32, error) {
	libhttpredirect, err := syscall.LoadLibrary("youniverse.dll")
	if err != nil {
		return 0, err
	}

	addrFuncation, err := syscall.GetProcAddress(libhttpredirect, "SetBusinessData")
	if err != nil {
		return 0, err
	}

	ret, _, _ := syscall.Syscall(addrFuncation, 2,
		uintptr(unsafe.Pointer(&addrPACSocket)),
		uintptr(unsafe.Pointer(&addrEncodeSocket)),
		0)

	syscall.FreeLibrary(syscall.Handle(libhttpredirect))

	return int32(ret), nil
}
