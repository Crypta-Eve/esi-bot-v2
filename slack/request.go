package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/eveisesi/eb2"
	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

func (s *service) makeESIDynamicRequestMessage(event Event) {

	parsed, valid := validateRoute(event.trigger)
	if !valid {

		attachment := nslack.Attachment{
			Pretext: "Provided Path is not valid. Please validate the path submitted and try again. The following is the final parsed route and what the engine used to validate this request.",
			Text:    parsed,
		}
		s.logger.Info("Responding to request for esi data.")
		channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionAttachments(attachment))
		if err != nil {
			s.logger.WithError(err).Error("failed to respond to request for esi data.")
			return
		}
		s.logger.WithFields(logrus.Fields{
			"channel":   channel,
			"timestamp": timestamp,
		}).Info("successfully responded to request for esi data")

		return

	}

	var base *url.URL
	var err error
	base, err = url.Parse(eb2.ESI_BASE)
	if err != nil {
		// TODO: Return Parsing Error
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), false))
	}

	uri, err := url.ParseRequestURI(parsed)
	if err != nil {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), false))

	}

	uri.Host = base.Host
	uri.Scheme = base.Scheme

	// Has nothing gone wrong yet? Amazing!!!
	start := time.Now()
	resp, err := http.Get(uri.String())
	if err != nil {
		// TODO: Error Handling for making the request.
		// This error does not throw if request.StatusCode != 200.
		// That is handled later
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), false))

	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), false))

	}

	if resp.StatusCode != 200 {
		txt := "The request to %s failed with status code %d and error message %s"

		_, _, _ = s.goslack.PostMessage(event.origin.Channel,
			nslack.MsgOptionText(
				fmt.Sprintf(
					txt,
					uri.String(),
					resp.StatusCode,
					string(data),
				),
				false,
			),
		)

	}

	dst := new(bytes.Buffer)
	_ = json.Indent(dst, data, "", "   ")

	data = dst.Bytes()

	if len(data) > 1010 {
		endtext := []byte("\nand more ...")
		data = data[:1010]
		data = append(data, endtext...)
	}

	d := time.Since(start)

	s.logger.Info("Responding to request for esi data.")
	_, err = s.goslack.UploadFile(nslack.FileUploadParameters{
		Filename:       "response.json",
		Filetype:       "json",
		Channels:       []string{event.origin.Channel},
		Content:        string(data),
		InitialComment: fmt.Sprintf("%s (%dms)", strings.ToUpper(resp.Status), d.Milliseconds()),
		Title:          uri.String(),
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to respond to request for esi data.")
		return
	}
	s.logger.Info("successfully responded to request for esi data")

}

func validateRoute(route string) (string, bool) {
	version := ""
	versions := []string{"latest", "legacy", "dev", "v1", "v2", "v3", "v4", "v5", "v6"}
	parsedCommand := strings.Split(strings.TrimSuffix(strings.TrimPrefix(route, "/"), "/"), "/")
	if strInStrSlice(parsedCommand[0], versions) {
		version = parsedCommand[0]
		parsedCommand = parsedCommand[1:]
	}
	validRoute := false

	for _, route := range rl {
		if len(route) != len(parsedCommand) {
			continue
		}
		for i, rpiece := range route {
			if strings.HasPrefix(rpiece, "{") && strings.HasSuffix(rpiece, "}") {
				if i == len(route)-1 {
					// The last index of the parsed route is a placeholder,
					//  lets exit
					validRoute = true
					goto Exit
				}
				continue
			}

			if route[i] == parsedCommand[i] {
				if i == len(route)-1 {
					// We are at the end of the route and everything has match,
					//  lets exit with the selected route
					validRoute = true
					goto Exit
				}
				// route matches so far, lets move onto the next rpiece
				continue
			}
			// This piece doesn't match the same piece of the parsed input,
			//  lets skip to the next route in the spec
			if route[i] != parsedCommand[i] {
				goto NextRoute
			}

		}
	NextRoute:
	}
Exit:

	if version == "" {
		version = "latest"
	}

	parsedCommand = append([]string{version}, parsedCommand...)

	return fmt.Sprintf("/%s", strings.Join(parsedCommand, "/")), validRoute
}
