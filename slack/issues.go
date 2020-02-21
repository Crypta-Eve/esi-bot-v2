package slack

import (
	"fmt"

	"github.com/eveisesi/eb2"
)

var newA = Attachment{
	Title: "Opening a new issue",
	Text:  fmt.Sprintf("Before opening a new issues, please use the <%s/issues|search function> to see if a similar issue exists, or has already been created.", eb2.ESI_ISSUES),
}

var br = Attachment{
	Title: "Report a New Bug",
	Text:  fmt.Sprintf("• unexpected 500 responses\n• incorrect information in the swagger spec\n• otherwise invalid or unexpected responses"),
	Color: "danger",
	Actions: []AttachmentAction{
		AttachmentAction{
			Type:  "button",
			Text:  "Report A Bug",
			URL:   fmt.Sprintf("%s/issues/new?template=bug.md", eb2.ESI_ISSUES),
			Style: "danger",
		},
	},
}

var fr = Attachment{
	Title: "Request A New Feature",
	Text:  "• adding an attribute to an existing route\n• exposing other readily available client data\n• meta requests, adding some global parameter to the specs",
	Color: "good",
	Actions: []AttachmentAction{
		AttachmentAction{
			Type:  "button",
			Text:  "Request A Feature",
			URL:   fmt.Sprintf("%s/issues/new?template=feature_request.md", eb2.ESI_ISSUES),
			Style: "primary",
		},
	},
}

var incon = Attachment{
	Title: "Report An Inconsistency",
	Text:  "• two endpoints returning slightly different names for the same attribute\n• attribute values are returned with different formats for different routes",
	Color: "warning",
	Actions: []AttachmentAction{
		AttachmentAction{
			Type:  "button",
			Text:  "Report An Inconsistency",
			URL:   fmt.Sprintf("%s/issues/new?template=inconsistency.md", eb2.ESI_ISSUES),
			Style: "primary",
		},
	},
}

func (s *service) makeIssues(parsed SlashCommand) (Msg, error) {

	var msg = Msg{
		Text:         "",
		ResponseType: "in_channel",
	}

	switch parsed.Text {
	case "bug", "br":
		msg.Attachments = []Attachment{
			newA, br,
		}
	case "feature", "fr", "enhancement":
		msg.Attachments = []Attachment{
			newA, fr,
		}
	case "inconsistency":
		msg.Attachments = []Attachment{
			newA, incon,
		}
	default:
		msg.Attachments = []Attachment{
			newA, br, fr, incon,
		}

	}

	return msg, nil

}
