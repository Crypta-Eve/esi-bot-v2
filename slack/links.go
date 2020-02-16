package slack

import (
	"encoding/json"
	"fmt"

	"github.com/eveisesi/eb2"
)

// This file contains the make commands that return links to help sites

func (s *service) makeLinkMessage(command string) ([]byte, error) {

	res := response{
		ResponseType: "in_channel",
	}

	switch command {
	case id, ids, ranges:
		res.UnfurlLinks = true
		res.Text = fmt.Sprintf(
			"ID Ranges References: \n\t\t%s\nAsset `location_id` reference:\n\t\t%s",
			eb2.ID_RANGES,
			fmt.Sprintf(
				"%s/docs/asset_location_id",
				eb2.ESI_DOCS,
			),
		)
	case source, repo:
		res.UnfurlLinks = true
		res.Text = eb2.SOURCE

	case faq:
		res.Text = fmt.Sprintf("%s/docs/FAQ", eb2.ESI_DOCS)
	case issues:
		res.UnfurlLinks = true
		res.Text = eb2.ESI_ISSUES
	case sso:
		res.UnfurlLinks = true
		res.Text = eb2.SSO_ISSUES
	}

	return json.Marshal(res)

}
