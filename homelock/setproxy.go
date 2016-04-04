package homelock

import (
	"syscall"
	"unsafe"

	"github.com/ssoor/winapi"
)

func SetPACProxy(autoConfigURL string) bool {
	var proxyInternetOptions [5]winapi.INTERNET_PER_CONN_OPTION
	var proxyInternetOptionList winapi.INTERNET_PER_CONN_OPTION_LIST

	proxyInternetOptions[0].Option = winapi.INTERNET_PER_CONN_AUTOCONFIG_URL
	proxyInternetOptions[0].Value = uint64(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(autoConfigURL))))

	proxyInternetOptions[1].Option = winapi.INTERNET_PER_CONN_AUTODISCOVERY_FLAGS
	proxyInternetOptions[1].Value = 0

	proxyInternetOptions[2].Option = winapi.INTERNET_PER_CONN_FLAGS
	proxyInternetOptions[2].Value = winapi.PROXY_TYPE_AUTO_PROXY_URL | winapi.PROXY_TYPE_DIRECT

	proxyInternetOptions[3].Option = winapi.INTERNET_PER_CONN_PROXY_BYPASS
	proxyInternetOptions[3].Value = uint64(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("<local>"))))

	proxyInternetOptions[4].Option = winapi.INTERNET_PER_CONN_PROXY_SERVER
	proxyInternetOptions[4].Value = 0

	proxyInternetOptionList.OptionError = 0
	proxyInternetOptionList.Connection = nil

	proxyInternetOptionList.Size = uint32(unsafe.Sizeof(proxyInternetOptionList))

	proxyInternetOptionList.OptionCount = 5
	proxyInternetOptionList.Options = (*winapi.INTERNET_PER_CONN_OPTION)(unsafe.Pointer(&proxyInternetOptions))

	return winapi.InternetSetOption(nil, winapi.INTERNET_OPTION_PER_CONNECTION_OPTION, &proxyInternetOptionList, proxyInternetOptionList.Size)
}
