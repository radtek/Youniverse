package internest

import (
	"net/http"

	"github.com/ssoor/webapi"
)

type URLNestedType int

type URLNestedAPI struct {
	url    string
	title  string
	typeof URLNestedType
}

const (
	NestedJump = iota
	NestedMoved
	NestedIFrame
)

func NewURLNestedAPI(nestedType URLNestedType, nestedTitle string, nestedURL string) *URLNestedAPI {
	return &URLNestedAPI{
		url:    nestedURL,
		title:  nestedTitle,
		typeof: nestedType,
	}
}

func (nested URLNestedAPI) Get(values webapi.Values, request *http.Request) (int, interface{}, http.Header) {

	html := `<!DOCTYPE html><html lang="zh-cmn-Hans">
				<head>
					<meta charset="gb2312"/><meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1" /><meta name="renderer" content="webkit" /><title>` + nested.title + `</title>
				</head>
				<body>
					<iframe src="` + nested.url + `" frameborder="0" style="overflow:hidden;overflow-x:hidden;overflow-y:hidden;height:100%;width:100%;position:absolute;top:0px;left:0px;right:0px;bottom:0px" height="100%" width="100%"></iframe> 
				</body>
			</html>`

	switch nested.typeof {
	case NestedJump:
		return http.StatusFound, []byte(html), http.Header{"Location": []string{nested.url}}
	case NestedMoved:
		return http.StatusMovedPermanently, []byte(html), http.Header{"Location": []string{nested.url}}
	case NestedIFrame:
		// 由于 默认IFrame，什么都不执行即可
	}

	return http.StatusOK, []byte(html), nil // 默认IFrame
}
