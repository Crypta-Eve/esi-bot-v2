package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/eveisesi/eb2"
	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

type StatusCategory struct {
	Status string
	Emoji  string
	Color  string
}

var categories = []StatusCategory{
	{
		Status: "red",
		Emoji:  ":fire:",
		Color:  "danger",
	},
	{
		Status: "yellow",
		Emoji:  ":fire_engine:",
		Color:  "warning",
	},
}

func (s *service) handleEveTQStatus(event Event) {
	s.makeEveServerStatusMessage(event, eb2.ESI_TRANQUILITY)
}

func (s *service) handleEveSerenityStatus(event Event) {
	s.makeEveServerStatusMessage(event, eb2.ESI_SERENITY)
}

func (s *service) makeEveServerStatusMessage(event Event, server string) {

	var base string
	var ok bool

	if base, ok = eb2.ESI_URLS[server]; !ok {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText("Unable to determine server to fetch status for", true))
		return
	}

	uri, _ := url.Parse(base)
	uri.Path = "/v1/status"

	resp, err := http.Get(uri.String())
	if err != nil {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), true))
		return
	}

	title := fmt.Sprintf("%s Status", strings.Title(server))

	defer resp.Body.Close()
	var attachment nslack.Attachment
	if resp.StatusCode > 200 {

		if resp.StatusCode != 503 {
			indeterminate := "Cannet Determine server status. It might be offline, or experiencing connectivity issues."
			attachment = nslack.Attachment{
				Color: "danger",
				// Leaving this like this so that we can support other servers in the future
				Title:    title,
				Text:     indeterminate,
				Fallback: fmt.Sprintf("%s Status: %s", "Tranquility", indeterminate),
			}

		} else {
			attachment = nslack.Attachment{
				Color: "danger",
				// Leaving this like this so that we can support other servers in the future
				Title:    title,
				Text:     "Offline",
				Fallback: fmt.Sprintf("%s Status: Offline", "Tranquility"),
			}
		}

		s.logger.Info("Responding to request for eve server status")
		channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionAttachments(attachment))
		if err != nil {
			s.logger.WithError(err).Error("failed to respond to request for eve server status.")
			return
		}
		s.logger.WithFields(logrus.Fields{
			"channel":   channel,
			"timestamp": timestamp,
		}).Info("successfully responded to request for eve server status")
		return

	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), true))
		return
	}
	var status eb2.ServerStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), true))
		return
	}
	color := "good"
	inVip := ""
	if status.Vip {
		color = "warning"
		inVip = ", in VIP"
	}
	attachment = nslack.Attachment{
		Color: color,
		Title: title,
		Fields: []nslack.AttachmentField{
			nslack.AttachmentField{
				Title: "Players Online",
				Value: humanize.Comma(status.Players),
			},
			nslack.AttachmentField{
				Title: "Started At",
				Value: status.StartTime.Format(layoutESI),
				Short: true,
			},
			nslack.AttachmentField{
				Title: "Running For",
				Value: determineServerRunTime(status.StartTime),
				Short: true,
			},
		},
		Fallback: fmt.Sprintf("%s status: %d player online, started at %s%s", "Tranquility", status.Players, status.StartTime.Format(layoutESI), inVip),
	}

	s.logger.Info("Responding to request for eve server status")
	channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionAttachments(attachment))
	if err != nil {
		s.logger.WithError(err).Error("failed to respond to request for eve server status.")
		return
	}
	s.logger.WithFields(logrus.Fields{
		"channel":   channel,
		"timestamp": timestamp,
	}).Info("successfully responded to request for eve server status")
}

func determineServerRunTime(from time.Time) string {

	n := time.Since(from)
	return time.Time{}.Add(n).Format("15h 04m 05s")

}

func (s *service) fetchRouteStatuses(version string) (routes []*eb2.ESIStatus, err error) {

	uri, _ := url.Parse(eb2.ESI_URLS[eb2.ESI_TRANQUILITY])
	uri.Path = "status.json"

	query := url.Values{}
	query.Set("version", version)

	uri.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	var currentEtag string
	etag, found := s.caches["etags"].Get(version)
	if found {
		currentEtag = etag.(string)
		req.Header.Add("If-None-Match", currentEtag)
	}

	// s.logger.WithField("current_etag", currentEtag).Debug()

	s.logger.WithField("req_etag", req.Header.Get("If-None-Match")).Debugln("Request Headers")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// s.logger.WithField("status_code", resp.StatusCode).Print()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unable to fetch route status. esi api responsed with an HTTP Status Code of %d", resp.StatusCode)
	}

	s.logger.WithField("res_etag", resp.Header.Get("Etag")).WithField("res_status", resp.Status).Debugln("Response Headers")

	// Emulate a 304 Response since this endpoint delivers back a 200 when there are no change, sometimes
	if currentEtag == resp.Header.Get("Etag") {
		return nil, nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &routes)
	if err != nil {
		return nil, err
	}

	s.caches["etags"].Flush()
	s.caches["etags"].Set(version, resp.Header.Get("Etag"), 0)

	return

}

func (s *service) handleESIStatusMessage(event Event) {
	version := "latest"
	if _, ok := event.flags["version"]; ok {
		version = event.flags["version"]
	}

	routes, found := s.checkCache(version)
	if !found {

		routes, err := s.fetchRouteStatuses(version)
		if err != nil {
			_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), true))
			return
		}

		s.caches["routes"].Flush()
		s.caches["routes"].Set(version, routes, 0)

	}

	s.MakeESIStatusMessage(event.origin.Channel, routes, version)

}

func (s *service) MakeESIMutatedRoutesMessage(channelID string, routes []string) {

	var attachments = []nslack.Attachment{
		{
			Color: "warning",
			Text:  fmt.Sprintf("```%s```", strings.Join(routes, "\n")),
		},
	}

	options := []nslack.MsgOption{}
	options = append(options, nslack.MsgOptionAttachments(attachments...))

	now := time.Now()
	msg := fmt.Sprintf("*ESI Schema Update Detected*\n\n<!date^%d^{date_num} {time_secs}|%s>", now.Unix(), now.Format("2006-01-02 15:04:05"))
	options = append(options, nslack.MsgOptionText(msg, false))

	channel, timestamp, err := s.goslack.PostMessage(channelID, options...)
	if err != nil {
		s.logger.WithError(err).Error("failed to send message about change in routes.")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"channel":   channel,
		"timestamp": timestamp,
	}).Info("successfully responded to request for esi route status.")

}

func (s *service) MakeESIStatusMessage(channelID string, routes []*eb2.ESIStatus, version string) {

	var etag string
	etagCheck, found := s.caches["etags"].Get(version)
	if found {
		etag = etagCheck.(string)
	}

	var attachments []nslack.Attachment
	for _, category := range categories {
		categoryRoutes := []*eb2.ESIStatus{}
		for _, route := range routes {
			if route.Status == category.Status {
				categoryRoutes = append(categoryRoutes, route)
			}
		}

		if len(categoryRoutes) > 0 {
			attachment := nslack.Attachment{
				Color: category.Color,
				Fallback: fmt.Sprintf(
					"%s: %d out of %d, %.3f%%",
					strings.Title(category.Status),
					len(categoryRoutes),
					len(routes),
					percentage(len(categoryRoutes), len(routes)),
				),
				Text: fmt.Sprintf(
					"%s %s %s %s",
					category.Emoji,
					fmt.Sprintf(
						"%d %s (out of %d,  %.3f%%)",
						len(categoryRoutes),
						strings.Title(category.Status),
						len(routes),
						percentage(len(categoryRoutes), len(routes)),
					),
					category.Emoji,
					generateRoutesString(categoryRoutes),
				),
			}
			attachments = append(attachments, attachment)
		}
	}
	if len(attachments) == 0 {
		attachments = append(attachments, nslack.Attachment{
			Text: ":the_horns:",
		})
	}
	now := time.Now()
	attachments[0].Pretext = fmt.Sprintf("<!date^%d^{date_num} {time_secs}|%s>", now.Unix(), now.Format("2006-01-02 15:04:05"))
	attachments[len(attachments)-1].Footer = fmt.Sprintf("Etag: %s\n", etag)

	options := []nslack.MsgOption{}
	options = append(options, nslack.MsgOptionAttachments(attachments...))
	if channelID != s.config.SlackESIStatusChannel {
		msg := fmt.Sprintf("Psst.....Checkout <#%s> for a continuous feed of statuses from me...", s.config.SlackESIStatusChannel)
		options = append(options, nslack.MsgOptionText(msg, false))
	}

	s.logger.Info("Responding to request for esi route status.")
	channel, timestamp, err := s.goslack.PostMessage(channelID, options...)
	if err != nil {
		s.logger.WithError(err).Error("failed to respond to request for esi route status.")
		return
	}
	s.logger.WithFields(logrus.Fields{
		"channel":   channel,
		"timestamp": timestamp,
	}).Info("successfully responded to request for esi route status.")

}

func (s *service) checkCache(version string) ([]*eb2.ESIStatus, bool) {
	check, found := s.caches["routes"].Get(version)
	if found {
		return check.([]*eb2.ESIStatus), found
	}
	return nil, found
}

func percentage(top int, bottom int) float64 {
	if bottom == 0 {
		return 0.00
	}
	return ((float64(top) / float64(bottom)) * 100)
}

func generateRoutesString(routes []*eb2.ESIStatus) string {

	if len(routes) == 0 {
		return ""
	}

	t := []string{}

	processed := 0
	for i, route := range routes {
		processed++
		t = append(t, fmt.Sprintf(
			"%s %s", strings.ToUpper(route.Method), route.Route,
		))
		if processed > 50 {
			t = append(t, fmt.Sprintf(
				"%d", len(routes[i:]),
			))
			break
		}

	}

	return fmt.Sprintf("```%s```", strings.Join(t, "\n"))
}
