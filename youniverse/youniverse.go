package youniverse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ssoor/groupcache"
	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/log"
)

var Resource *groupcache.Group

var (
	ErrorYouniverseUninit = errors.New("youniverse not initialization")
)

func Get(ctx groupcache.Context, key string, dest *[]byte) error {
	if nil == Resource {
		return ErrorYouniverseUninit
	}

	defer log.TimeoutWarning(fmt.Sprint("Youniverse get resource ", key), time.Now(), 5)

	return Resource.Get(ctx, key, groupcache.AllocatingByteSliceSink(dest))
}

func getPeers(guid string, peer_addr string) ([]string, error) {
	url := "http://120.26.80.61/issued/peers/20160404/" + guid + ".peers?peer=" + peer_addr

	json_peers, err := api.GetURL(url)
	if err != nil {
		return []string{}, errors.New("Query peers interface failed.")
	}

	peers := []string{}
	if err = json.Unmarshal([]byte(json_peers), &peers); err != nil {
		return []string{}, errors.New("Unmarshal peers interface failed.")
	}

	return peers, nil
}

// DefaultTransport is the default implementation of Transport and is
// used by DefaultClient. It establishes network connections as needed
// and caches them for reuse by subsequent calls. It uses HTTP proxies
// as directed by the $HTTP_PROXY and $NO_PROXY (or $http_proxy and
// $no_proxy) environment variables.
// DefaultTransport is the default implementation of Transport and is
// used by DefaultClient. It establishes network connections as needed
// and caches them for reuse by subsequent calls. It uses HTTP proxies
// as directed by the $HTTP_PROXY and $NO_PROXY (or $http_proxy and
// $no_proxy) environment variables.
var GCHTTPPoolOptions *groupcache.HTTPPoolOptions = &groupcache.HTTPPoolOptions{
	BasePath: "youniverse",
	Transport: func(context groupcache.Context) http.RoundTripper {
		return &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	},
}

func StartYouniverse(guid string, peerAddr string, setting Settings) error {

	peers := groupcache.NewHTTPPoolOpts("http://"+peerAddr, GCHTTPPoolOptions)
	log.Info("Create Youiverse HTTP pool: http://" + peerAddr)

	peerUrls, err := getPeers(guid, "http://"+peerAddr)
	if nil != err {
		return err
	}

	log.Info("Set Youiverse peer:", len(peerUrls), peerUrls)

	for _, peerUrl := range peerUrls {
		peers.AddPeer(peerUrl)
	}

	client := NewBackend(setting.ResourceURLs)
	log.Info("Set Youiverse backend interfase:", setting.ResourceURLs)

	Resource = groupcache.NewGroup("resource", setting.MaxSize, groupcache.GetterFunc(
		func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
			dest.SetBytes([]byte(client.Get(key)))
			return nil
		}))

	go http.ListenAndServe(peerAddr, http.HandlerFunc(peers.ServeHTTP))

	return nil
}
