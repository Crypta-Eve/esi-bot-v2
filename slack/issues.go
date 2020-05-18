package slack

import (
	"fmt"

	"github.com/eveisesi/eb2"
	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

var header = nslack.Attachment{
	Title: "Opening a new issue",
	Text:  fmt.Sprintf("Before opening a new issues, please use the <%s/issues|search function> to see if a similar issue exists, or has already been created.", eb2.ESI_ISSUES),
}

var br = nslack.Attachment{
	Title: "Report a New Bug",
	Text:  "• unexpected 500 responses\n• incorrect information in the swagger spec\n• otherwise invalid or unexpected responses",
	Color: "danger",
	Actions: []nslack.AttachmentAction{
		nslack.AttachmentAction{
			Type:  "button",
			Text:  "Report A Bug",
			URL:   fmt.Sprintf("%s/issues/new?template=bug.md", eb2.ESI_ISSUES),
			Style: "danger",
		},
	},
}

var fr = nslack.Attachment{
	Title: "Request A New Feature",
	Text:  "• adding an attribute to an existing route\n• exposing other readily available client data\n• meta requests, adding some global parameter to the specs",
	Color: "good",
	Actions: []nslack.AttachmentAction{
		nslack.AttachmentAction{
			Type:  "button",
			Text:  "Request A Feature",
			URL:   fmt.Sprintf("%s/issues/new?template=feature_request.md", eb2.ESI_ISSUES),
			Style: "primary",
		},
	},
}

var incon = nslack.Attachment{
	Title: "Report An Inconsistency",
	Text:  "• two endpoints returning slightly different names for the same attribute\n• attribute values are returned with different formats for different routes",
	Color: "warning",
	Actions: []nslack.AttachmentAction{
		nslack.AttachmentAction{
			Type:  "button",
			Text:  "Report An Inconsistency",
			URL:   fmt.Sprintf("%s/issues/new?template=inconsistency.md", eb2.ESI_ISSUES),
			Style: "primary",
		},
	},
}

func (s *service) makeIssues(event Event) {

	var attachments []nslack.Attachment

	switch event.trigger {
	case "bug", "br":
		attachments = []nslack.Attachment{
			header, br,
		}
	case "feature", "fr", "enhancement":
		attachments = []nslack.Attachment{
			header, fr,
		}
	case "inconsistency":
		attachments = []nslack.Attachment{
			header, incon,
		}
	default:
		attachments = []nslack.Attachment{
			header, br, fr, incon,
		}
	}

	s.logger.Info("Responding to issues request")
	channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionAttachments(attachments...))
	if err != nil {
		s.logger.WithError(err).Error("failed to respond to request for help.")
		return
	}
	s.logger.WithFields(logrus.Fields{
		"channel":   channel,
		"timestamp": timestamp,
	}).Info("successfully responded issues request")

}
