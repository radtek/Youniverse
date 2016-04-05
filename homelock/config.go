package homelock

import (
	"github.com/ssoor/youniverse/homelock/socksd"
)

type Settings struct {
	Encode   bool              `json:"encode"`
	Services []socksd.Upstream `json:"services"`
}
