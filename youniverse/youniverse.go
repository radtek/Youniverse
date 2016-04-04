package youniverse

import (
	"net/http"
	"strconv"

	"github.com/ssoor/groupcache"
	"github.com/ssoor/youniverse/log"
)

var cache *groupcache.Group

func StartYouniverse(port int16,max_size int64, peer_urls []string) {
	backendURLs := []string{"http://localhost/youniverse/resource"}

	peers := groupcache.NewHTTPPool("http://localhost:" + strconv.Itoa(port))
	log.Info.Println("Create Youiverse HTTP pool: http://localhost:" + strconv.Itoa(port))

	log.Info.Println("Set Youiverse peer:")
	for _, peer_url := range peer_urls {
		peers.AddPeer(peer_url)
		log.Info.Printf("\t%s", peer_url)
	}

	client := NewBackend(backendURLs)
	log.Info.Println("Set Youiverse backend interfase:", backendURLs)

	cache = groupcache.NewGroup("resource", max_size, groupcache.GetterFunc(
		func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
			dest.SetBytes([]byte(client.Get(key)))
			return nil
		}))

	go http.ListenAndServe("localhost:"+strconv.Itoa(port), http.HandlerFunc(peers.ServeHTTP))
}
