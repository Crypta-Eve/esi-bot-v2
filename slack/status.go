package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/eveisesi/eb2"
	"github.com/patrickmn/go-cache"
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

var statusCache = cache.New(time.Minute*1, time.Second*30)

func (s *service) makeEveServerStatusMessage(parsed SlashCommand) (Msg, error) {

	msg := Msg{}

	uri, _ := url.Parse(eb2.ESI_BASE)
	uri.Path = "/v1/status"

	resp, err := http.Get(uri.String())
	if err != nil {
		return Msg{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 200 {

		if resp.StatusCode != 503 {
			indeterminate := "Cannet Determine server status. It might be offline, or experiencing connectivity issues."
			msg.Attachments = []Attachment{
				Attachment{
					Color: "danger",
					// Leaving this like this so that we can support other servers in the future
					Title:    fmt.Sprintf("%s status", "Tranquility"),
					Text:     indeterminate,
					Fallback: fmt.Sprintf("%s Status: %s", "Tranquility", indeterminate),
				},
			}

			return msg, nil
		}

		msg.Attachments = []Attachment{
			Attachment{
				Color: "danger",
				// Leaving this like this so that we can support other servers in the future
				Title:    fmt.Sprintf("%s status", "Tranquility"),
				Text:     "Offline",
				Fallback: fmt.Sprintf("%s Status: Offline", "Tranquility"),
			},
		}

	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Msg{}, err
	}
	var status eb2.ServerStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return Msg{}, err
	}
	color := "good"
	inVip := ""
	if status.Vip {
		color = "warning"
		inVip = ", in VIP"
	}
	msg.Attachments = []Attachment{
		Attachment{
			Color: color,
			Title: fmt.Sprintf("%s status", "Tranquility"),
			Fields: []AttachmentField{
				AttachmentField{
					Title: "Players Online",
					Value: fmt.Sprintf("%d", status.Players),
				},
				AttachmentField{
					Title: "Started At",
					Value: status.StartTime.Format(layoutESI),
					Short: true,
				},
				AttachmentField{
					Title: "Running For",
					Value: determineServerRunTime(status.StartTime),
					Short: true,
				},
			},
			Fallback: fmt.Sprintf("%s status: %d player online, started at %s%s", "Tranquility", status.Players, status.StartTime.Format(layoutESI), inVip),
		},
	}

	return msg, err
}

func determineServerRunTime(from time.Time) string {

	n := time.Since(from)
	return time.Time{}.Add(n).Format("15h 04m 05s")

}

func (s *service) makeESIStatusMessage(parsed SlashCommand) (Msg, error) {

	uri, _ := url.Parse(eb2.ESI_BASE)
	uri.Path = "status.json"

	version := "latest"
	if _, ok := parsed.Args["version"]; ok {
		version = parsed.Args["version"]
	}

	routes, found := checkCache(version)
	s.logger.WithField("cache_hit", found).Info("processing request for esi route status")
	if !found {

		query := url.Values{}
		query.Set("version", version)

		uri.RawQuery = query.Encode()

		resp, err := http.Get(uri.String())
		if err != nil {
			return Msg{}, err
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return Msg{}, err
		}

		err = json.Unmarshal(data, &routes)
		if err != nil {
			return Msg{}, err
		}

		statusCache.Flush()
		statusCache.Set(version, routes, 0)

	}

	var attachments []Attachment
	for _, category := range categories {
		categoryRoutes := []eb2.ESIStatus{}
		for _, route := range routes {
			if route.Status == category.Status {

				categoryRoutes = append(categoryRoutes, route)
			}
		}

		if len(categoryRoutes) > 0 {

			attachment := Attachment{
				Color: category.Color,
				Fallback: fmt.Sprintf(
					"%s: %d out of %d, %.3f",
					strings.Title(category.Status),
					len(categoryRoutes),
					len(routes),
					percentage(len(categoryRoutes), len(routes)),
				),
				Text: fmt.Sprintf(
					"%s %s %s %s",
					category.Emoji,
					fmt.Sprintf(
						"%d %s (out of %d,  %.3f)",
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
		attachments = append(attachments, Attachment{
			Text: ":ok_hand:",
		})
	}

	return Msg{
		Attachments: attachments,
	}, nil

}

func checkCache(version string) ([]eb2.ESIStatus, bool) {
	check, found := statusCache.Get(version)
	if found {
		return check.([]eb2.ESIStatus), true
	}
	return []eb2.ESIStatus{}, false
}

func percentage(top int, bottom int) float64 {
	if bottom == 0 {
		return 0.00
	}
	return 1 - (float64(top) / float64(bottom) * 100)
}

func generateRoutesString(routes []eb2.ESIStatus) string {

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
