package pac

import (
	"bytes"
	"net/http"

	"github.com/ssoor/youniverse/log"
)

func StartPACServer(pac PAC) {
	pu, err := NewPACUpdater(pac)
	if err != nil {
		log.Error("failed to NewPACUpdater, err:", err)
		return
	}

	http.HandleFunc("/proxy.pac", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-ns-proxy-autoconfig")
		data, time := pu.get()
		reader := bytes.NewReader(data)
		http.ServeContent(w, r, "proxy.pac", time, reader)
	})

	err = http.ListenAndServe(pac.Address, nil)

	if err != nil {
		log.Error("listen failed, err:", err)
		return
	}
}
