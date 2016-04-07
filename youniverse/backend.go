package youniverse

// Client for dbserver/slowdb

import (
	"io/ioutil"
	"net/http"
    
    "github.com/ssoor/youniverse/log"
)

type Backend struct {
	baseURLs []string
}

func NewBackend(base []string) Backend {
	return Backend{
		baseURLs: base,
	}
}

func (b *Backend) Get(key string) []byte {
	var data []byte

	for _, baseURL := range b.baseURLs {
		resp, err := http.Get(baseURL + key)
		if err != nil {
			log.Warning(err)
			continue
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Warning("Backend download resource",baseURL + key,"failed, interface result stats:", resp.StatusCode)
			continue
		}

		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			continue
		}

		break
	}

	return data
}
