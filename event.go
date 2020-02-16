package eb2

type EventWrapper struct {
	Token     string `json:"token"`
	TeamID    string `json:"team_id"`
	APIAppID  string `json:"api_app_id"`
	Event     Event  `json:"event"`
	Type      string `json:"type"`
	EventID   string `json:"event_id"`
	EventTime uint64 `json:"event_time"`

	// Unique Field that is only present during initial config of the app
	Challenge string `json:"challenge"`
}

type Event struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	User        string `json:"user"`
	Team        string `json:"team"`
	Channel     string `json:"channel"`
	ChannelType string `json:"channel_type"`
}
