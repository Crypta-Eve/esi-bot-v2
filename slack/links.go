package slack

import (
	"fmt"

	"github.com/eveisesi/eb2"
)

// This file contains the make commands that return links to help sites

func (s *service) makeLinkMessage(parsed SlashCommand) (Msg, error) {

	res := Msg{
		UnfurlLinks: true,

		ResponseType: "in_channel",
	}

	switch parsed.Text {
	case "id", "ids", "ranges":
		res.Text = fmt.Sprintf(
			"ID Ranges References: \n\t\t%s\nAsset `location_id` reference:\n\t\t%s",
			eb2.ID_RANGES,
			fmt.Sprintf(
				"%s/docs/asset_location_id",
				eb2.ESI_DOCS,
			),
		)
	case "source", "repo":
		res.Text = fmt.Sprintf("I'm an open source bot. If you want to contribute or you're curious how I work, my source is available for you to browse here: %s", eb2.SOURCE)
	case "faq":
		res.Text = fmt.Sprintf("%s/docs/FAQ", eb2.ESI_DOCS)
	case "issues":
		res.Text = fmt.Sprintf("%s/issues", eb2.ESI_ISSUES)
	case "sso":
		res.Text = fmt.Sprintf("%s/issues", eb2.SSO_ISSUES)
	}

	return res, nil

}
