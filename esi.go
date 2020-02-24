package eb2

import "time"

type ESIStatus struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Route    string `json:"route"`
	Status   string `json:"status"`
}

type ServerStatus struct {
	Players       int64     `json:"players"`
	ServerVersion string    `json:"server_version"`
	StartTime     time.Time `json:"start_time"`
	Vip           bool      `json:"vip"`
}
