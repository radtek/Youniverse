package internest

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ssoor/webapi"
	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/log"
)

type WarrantAPI struct {
	deviceCode    string
	warrantCode   string
	signBussiness map[string]string
}

func NewWarrantAPI(deviceCode string, warrantCode string) *WarrantAPI {
	return &WarrantAPI{
		deviceCode:    deviceCode,
		warrantCode:   warrantCode,
		signBussiness: make(map[string]string),
	}
}

func (this WarrantAPI) Get(values webapi.Values, request *http.Request) (int, interface{}, http.Header) {
	bussinessType := values.Get("type")

	if 0 == len(bussinessType) {
		return 501, nil, nil
	}

	if _, isHave := this.signBussiness[bussinessType]; isHave {
		return 200, this.signBussiness[bussinessType], nil
	}

	url, err := url.Parse("http://social.ssoor.com/warrant/enforce/20160308/" + this.warrantCode + ".enforce")
	if nil != err {
		return 501, err, nil
	}

	url.RawQuery = values.Encode()

	log.Info("Internest enforce bussiness:", url)

	json_sign, err := api.GetURL(url.String())
	if err != nil {
		return 501, errors.New("Query internest enforce interface failed."), nil
	}

	this.signBussiness[bussinessType] = json_sign

	return 200, json_sign, nil
}
