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
	return http.StatusFound, []byte("Internal error..."), http.Header{"Location": []string{api.url}}
}
