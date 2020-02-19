package slack

import (
	"fmt"
)

func (s *service) makeGreeting(parsed SlashCommand) (Msg, error) {

	user := parsed.UserID

	res := Msg{
		ResponseType: "in_channel",
	}

	switch parsed.Text {
	case "o7", "o/":
		res.Text = fmt.Sprintf("o7 <@%s>", user)
	case "7o", "\\o":
		res.Text = fmt.Sprintf("7o <@%s>", user)
	default:
		res.Text = fmt.Sprintf("hey <@%s>, hows it goin?", user)
	}

	return res, nil
}
