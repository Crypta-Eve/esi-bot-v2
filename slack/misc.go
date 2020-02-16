package slack

import (
	"encoding/json"
	"fmt"
)

func (s *service) makeHelloMessage(user, message string) ([]byte, error) {
	text := ""

	switch message {
	case "o7", "o/":
		text = fmt.Sprintf("o7 <@%s>", user)
	case "7o", "\\o":
		text = fmt.Sprintf("7o <@%s>", user)
	default:
		text = fmt.Sprintf("hey <@%s>, hows it goin?", user)
	}

	return json.Marshal(response{
		ResponseType: "in_channel",
		Text:         text,
	})
}
