//nolint:gochecknoglobals// who cares?
package slack

import "api/config/env"

type Channel struct {
	Local string
	Dev   string
	Prod  string
}

var StatusChannel = &Channel{
	Local: "C07JD83V10S",
	Dev:   "C082950MMFU",
	Prod:  "C07JEDT4LPN",
}

func (c *Channel) ToString(e env.Env) string {
	switch e {
	case env.Local:
		return c.Local
	case env.Dev:
		return c.Dev
	case env.Prod:
		return c.Prod
	default:
		return c.Prod
	}
}
