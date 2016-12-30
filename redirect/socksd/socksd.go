package socksd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/log"
)

type UpstreamDialer struct {
	url string

	index   int
	lock    sync.Mutex
	dialers []socks.Dialer
}

func NewUpstreamDialer(url string) *UpstreamDialer {
	dialer := &UpstreamDialer{
		url:     url,
		dialers: []socks.Dialer{NewDecorateDirect(0)}, // 原始连接,不经过任何处理
	}

	go dialer.backgroundUpdateServices()

	return dialer
}

func getSocksdSetting(url string) (setting Setting, err error) {
	var jsonData string

	if jsonData, err = api.GetURL(url); err != nil {
		return setting, errors.New(fmt.Sprint("Query setting interface failed, err: ", err))
	}

	if err = json.Unmarshal([]byte(jsonData), &setting); err != nil {
		return setting, errors.New("Unmarshal setting interface failed.")
	}

	return setting, nil
}

func buildUpstream(upstream Upstream, forward socks.Dialer) (socks.Dialer, error) {
	cipherDecorator := NewCipherConnDecorator(upstream.Crypto, upstream.Password)
	forward = NewDecorateClient(forward, cipherDecorator)

	switch strings.ToLower(upstream.Type) {
	case "socks5":
		{
			return socks.NewSocks5Client("tcp", upstream.Address, "", "", forward)
		}
	case "shadowsocks":
		{
			return socks.NewShadowSocksClient("tcp", upstream.Address, forward)
		}
	}
	return nil, errors.New("unknown upstream type" + upstream.Type)
}

func buildSetting(setting Setting) []socks.Dialer {
	var allForward []socks.Dialer
	for _, upstream := range setting.Upstreams {
		var forward socks.Dialer
		var err error
		forward = NewDecorateDirect(setting.DNSCacheTime)
		forward, err = buildUpstream(upstream, forward)
		if err != nil {
			log.Warning("failed to BuildUpstream, err:", err)
			continue
		}
		allForward = append(allForward, forward)
	}
	if len(allForward) == 0 {
		router := NewDecorateDirect(setting.DNSCacheTime)
		allForward = append(allForward, router)
	}

	return allForward
}

func (u *UpstreamDialer) backgroundUpdateServices() {
	var err error
	var setting Setting

	for {
		if setting, err = getSocksdSetting(u.url); nil != err {
			continue
		}

		log.Info("Setting messenger server information:")
		for _, upstream := range setting.Upstreams {
			log.Info("\tUpstream :", upstream.Address)
		}
		log.Info("\tDNS cache timeout time :", setting.DNSCacheTime)
		log.Info("\tNext flush interval time :", setting.IntervalTime)

		u.lock.Lock()
		u.index = 0
		u.dialers = buildSetting(setting)
		u.lock.Unlock()

		time.Sleep(time.Duration(setting.IntervalTime) * time.Second)
	}
}

func (u *UpstreamDialer) getNextDialer() socks.Dialer {
	u.lock.Lock()
	defer u.lock.Unlock()
	index := u.index
	u.index++
	if u.index >= len(u.dialers) {
		u.index = 0
	}
	if index < len(u.dialers) {
		return u.dialers[index]
	}
	panic("unreached")
}

func (u *UpstreamDialer) Dial(network, address string) (net.Conn, error) {
	router := u.getNextDialer()
	conn, err := router.Dial(network, address)
	if err != nil {
		log.Error("UpstreamDialer router.Dial failed, err:", err, network, address)
		return nil, err
	}
	return conn, nil
}
