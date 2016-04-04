package backend

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

func (b *Backend) Get(key string) string {
	var data []byte

	for _, baseURL := range b.baseURLs {
		resp, err := http.Get(baseURL + key)
		if err != nil {
			log.Warning.Printf("%s", err)
			continue
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Warning.Printf("%s", err)
			continue
		}

		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error.Printf("%s", err)
			continue
		}

		break
	}

	return string(data)
}
