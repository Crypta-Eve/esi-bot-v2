package slack

import (
	"fmt"
	"strings"
)

func (s *service) makeHelpMessage(parsed SlashCommand) (Msg, error) {

	res := Msg{}

	var blob []string

	catLen := len(s.commands)

	for i, category := range s.commands {

		blob = append(blob, fmt.Sprintf("%s - (%s):\n", category.Name, category.Description))

		for _, command := range category.Commands {

			trigger := command.Triggers
			description := command.Description

			if !command.Strict {
				trigger = []string{command.Example}
			}

			text := fmt.Sprintf("\t%s\n\t\t%s", strings.Join(trigger, ", "), description)

			text = fmt.Sprintf("%s\n", text)

			blob = append(blob, text)
		}

		if i != catLen-1 {
			blob = append(blob, "\n")
		}

	}

	res.Text = fmt.Sprintf("```%s```", strings.Join(blob, ""))

	return res, nil

}
