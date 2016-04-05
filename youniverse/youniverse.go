package youniverse

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ssoor/groupcache"
	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/log"
)

var cache *groupcache.Group

func getPeers(guid string,peer_addr string) ([]string, error) {
	url := "http://120.26.80.61/issued/peers/20160308/" + guid + ".peers?peer=" + peer_addr

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

func StartYouniverse(guid string, peerAddr string, setting Settings) error {

	peers := groupcache.NewHTTPPool("http://" + peerAddr)
	log.Info.Println("Create Youiverse HTTP pool: http://" + peerAddr)

	log.Info.Println("Set Youiverse peer:")
    peerUrls,err := getPeers(guid,peerAddr)
    
    if nil != err{
        return err
    }
    
	for _, peerUrl := range peerUrls {
		peers.AddPeer(peerUrl)
		log.Info.Printf("\t%s", peerUrl)
	}

	client := NewBackend(setting.ResourceURLs)
	log.Info.Println("Set Youiverse backend interfase:", setting.ResourceURLs)

	cache = groupcache.NewGroup("resource", setting.MaxSize, groupcache.GetterFunc(
		func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
			dest.SetBytes([]byte(client.Get(key)))
			return nil
		}))

	go http.ListenAndServe(peerAddr, http.HandlerFunc(peers.ServeHTTP))

	return nil
}
