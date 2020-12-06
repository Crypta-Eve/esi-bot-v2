package slack

import (
	"fmt"

	"github.com/eveisesi/eb2"
	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

// This file contains the make commands that return links to help sites

func (s *service) makeLinkMessage(event Event) {

	text := ""

	switch event.trigger {
	case "id", "ids", "ranges":
		text = fmt.Sprintf(
			"ID Ranges References: \n\t\t%s\nAsset `location_id` reference:\n\t\t%s",
			fmt.Sprintf(
				"%s/docs/id_ranges",
				eb2.ESI_DOCS,
			),
			fmt.Sprintf(
				"%s/docs/asset_location_id",
				eb2.ESI_DOCS,
			),
		)
	case "source", "repo":
		text = fmt.Sprintf("I'm an open source bot. If you want to contribute or you're curious how I work, my source is available for you to browse here: %s", eb2.SOURCE)
	case "faq":
		text = fmt.Sprintf("%s/docs/FAQ", eb2.ESI_DOCS)
	case "issues":
		text = fmt.Sprintf("%s/issues", eb2.ESI_ISSUES)
	case "sso":
		text = fmt.Sprintf("%s/issues", eb2.SSO_ISSUES)
	case "webui", "ui":
		text = fmt.Sprintf("%s/ui", eb2.ESI_URLS[eb2.ESI_TRANQUILITY])
	case "diff", "diffs":
		text = fmt.Sprintf("%s/diff/latest/dev", eb2.ESI_URLS[eb2.ESI_TRANQUILITY])
	case "ask":
		text = fmt.Sprintf("I have been summoned to ask you to read this important link if you want any further help! %s", eb2.ASK_TO_ASK)
	}

	s.logger.Info("Responding to a request for a link")
	channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(text, false), nslack.MsgOptionPostMessageParameters(nslack.PostMessageParameters{
		UnfurlLinks: true,
	}))
	if err != nil {
		s.logger.WithError(err).Error("failed to a request for a link.")
		return
	}
	s.logger.WithFields(logrus.Fields{
		"channel":   channel,
		"timestamp": timestamp,
	}).Info("successfully responded to a request for a link")

}
