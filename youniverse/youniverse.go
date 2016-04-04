package youniverse

import (
	"errors"
	"net"
	"net/http"
	"strconv"

	"github.com/ssoor/groupcache"
	"github.com/ssoor/youniverse/log"
)

var cache *groupcache.Group

var (
	ErrorSocketUnavailable error = errors.New("socket port not find")
)

func SocketSelectPort(port_type string, port_base int) (int16, error) {

	for ; port_base < 65536; port_base++ {

		tcpListener, err := net.Listen(port_type, ":"+strconv.Itoa(port_base))

		if err == nil {
			tcpListener.Close()
			return int16(port_base), nil
		}
	}

	return 0, ErrorSocketUnavailable
}

func StartYouniverse(port int16, max_size int64, peer_urls []string) error {
	port, err := SocketSelectPort("tcp", int(port))

	if err != nil {
		return err
	}

	http_addr := "localhost:" + strconv.Itoa(int(port))
	backendURLs := []string{"http://localhost/youniverse/resource/"}

	peers := groupcache.NewHTTPPool("http://" + http_addr)
	log.Info.Println("Create Youiverse HTTP pool: http://" + http_addr)

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

	go http.ListenAndServe(http_addr, http.HandlerFunc(peers.ServeHTTP))
    
    return nil
}
