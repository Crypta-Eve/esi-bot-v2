package slack

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (s *service) makeHelpMessage(recognize bool) ([]byte, error) {
	text := "The following commands are enabled: "
	if !recognize {
		text = "Hmmmm...that is a not a recognized command. Here is a list of commands that are enabled: "
	}

	message := response{
		ResponseType: "in_channel",
		Text:         fmt.Sprintf("%s\n`%s`", text, strings.Join(commands, "` `")),
	}

	return json.Marshal(message)

}
