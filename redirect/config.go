package redirect

import (
	"github.com/ssoor/youniverse/redirect/socksd"
)

type Settings struct {
	PAC   bool              `json:"pac"`
	Encode   bool              `json:"encode"`
	RulesURL string            `json:"rules_url"`
	BricksURL string            `json:"bricks_url"`
	Services []socksd.Upstream `json:"services"`
}
