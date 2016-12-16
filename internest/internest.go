package internest

import (
	"github.com/ssoor/webapi"
	"github.com/ssoor/youniverse/assistant"
	"github.com/ssoor/youniverse/common"
	"github.com/ssoor/youniverse/log"
)

func StartInternest(account string, guid string, setting Settings) (bool, error) {

	service := webapi.NewByteAPI()

	statsAPI := NewStatsAPI() // 程序运行状态
	service.AddResource(statsAPI, "/stats")

	for _, htmlNested := range setting.HtmlNested {
		htmlNestedAPI := NewHtmlNestedAPI(htmlNested.Status, []byte(htmlNested.Data), htmlNested.Header)
		service.AddResource(htmlNestedAPI, htmlNested.Path)
		log.Info("Sign resource html nested:", htmlNested.Path, ", status code is", htmlNested.Status, ", http header:", htmlNested.Header)
	}

	for _, urlNested := range setting.URLNested {
		urlnestedAPI := NewURLNestedAPI(URLNestedType(urlNested.Type), urlNested.Title, urlNested.URL, urlNested.ScriptURL)
		service.AddResource(urlnestedAPI, urlNested.Path)
		log.Info("Sign resource url nested", urlNested.Type, ":<", urlNested.Title, ">", urlNested.Path, ", url is", urlNested.URL, ", script is:", urlNested.ScriptURL)
	}

	if 0 == setting.APIPort {
		selectPort, err := common.SocketSelectPort("tcp")
		if nil != err {
			return false, err
		}

		setting.APIPort = int(selectPort)
	}

	go service.Start(setting.APIPort)

	handle, err := assistant.SetAPIPort2(setting.APIPort)
	log.Info("Setting internest", setting.APIPort, ", data share handle:", handle, ", err:", err)

	return true, nil
}
