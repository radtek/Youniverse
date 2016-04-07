package homelock

import (
	"github.com/ssoor/youniverse/homelock/socksd"
)

type Settings struct {
	Encode   bool              `json:"encode"`
	RulesURL string            `json:"rules_url"`
	BricksURL string            `json:"bricks_url"`
	Services []socksd.Upstream `json:"services"`
}
