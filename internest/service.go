package internest

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/ssoor/webapi"
	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/log"
)

type WarrantAPI struct {
	deviceCode    string
	enforceURL    string
	signBussiness map[string]string
}

func NewWarrantAPI(deviceCode string, enforceURL string) *WarrantAPI {
	return &WarrantAPI{
		enforceURL:    enforceURL,
		deviceCode:    deviceCode,
		signBussiness: make(map[string]string),
	}
}

func (this WarrantAPI) Get(values webapi.Values, request *http.Request) (int, interface{}, http.Header) {
	bussinessType := values.Get("type")

	if 0 == len(bussinessType) {
		return 501, nil, nil
	}

	if _, isHave := this.signBussiness[bussinessType]; isHave {
		return 200, []byte(this.signBussiness[bussinessType]), nil
	}

	url, err := url.Parse(this.enforceURL)
	if nil != err {
		return 501, err, nil
	}

	values.Set("internest", strconv.Itoa(os.Getpid()))
	
	url.RawQuery = values.Encode()
	log.Info("Internest enforce bussiness:", url)

	json_sign, err := api.GetURL(url.String())
	if err != nil {
		return 501, errors.New("Query internest enforce interface failed."), nil
	}

	this.signBussiness[bussinessType] = json_sign

	return 200, []byte(json_sign), nil
}
