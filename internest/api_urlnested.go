package internest

import (
	"net/http"

	"github.com/ssoor/webapi"
)

type URLNestedType int

type URLNestedAPI struct {
	url    string
	typeof URLNestedType
}

const (
	NestedJump = iota
	NestedMoved
	NestedIFrame
)

func NewURLNestedAPI(nestedType URLNestedType, nestedURL string) *URLNestedAPI {
	return &URLNestedAPI{
		url:    nestedURL,
		typeof: nestedType,
	}
}

func (api URLNestedAPI) Get(values webapi.Values, request *http.Request) (int, interface{}, http.Header) {

	html := `<!DOCTYPE html><html lang="zh-CN"><head><meta charset="gb2312"/><title>hao123_上网从这里开始</title></head><body><iframe src="` + api.url + `" frameborder="0" style="overflow:hidden;overflow-x:hidden;overflow-y:hidden;height:100%;width:100%;position:absolute;top:0px;left:0px;right:0px;bottom:0px" height="100%" width="100%"></iframe></body></html>`

	switch api.typeof {
	case NestedJump:
		return http.StatusFound, []byte(html), http.Header{"Location": []string{api.url}}
	case NestedMoved:
		return http.StatusMovedPermanently, []byte(html), http.Header{"Location": []string{api.url}}
	case NestedIFrame:
		// 由于 默认IFrame，什么都不执行即可
	}

	return http.StatusOK, []byte(html), nil // 默认IFrame
}
