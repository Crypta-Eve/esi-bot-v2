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
)

func (s *service) makeESIDynamicRequestMessage(parsed SlashCommand) (Msg, error) {

	s.logger.WithField("path", parsed.Text).Info("received ESI Path Request")

	parsedCommand := strings.Split(strings.TrimSuffix(strings.TrimPrefix(parsed.Text, "/"), "/"), "/")

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
	p := ""
	if validRoute {
		currentVersion := parsedCommand[0]
		hasVersion := false
		for _, v := range []string{"latest", "dev", "legacy"} {
			if v == currentVersion {
				hasVersion = true
			}
		}
		if !hasVersion {
			parsedCommand = append([]string{"latest"}, parsedCommand...)
		}

		p = fmt.Sprintf("/%s/", strings.Join(parsedCommand, "/"))
	} else {
		return Msg{
			ResponseType: "in_channel",
			Text:         "Provided Path is not valid. Please validate the path submitted and try again.",
			Attachments: []Attachment{
				Attachment{
					Pretext: "The following is the final parsed route and what the engine used to validate this request.",
					Text:    fmt.Sprintf("/%s/", strings.Join(parsedCommand, "/")),
				},
			},
		}, nil

	}

	var base *url.URL
	var err error
	base, err = url.Parse(eb2.ESI_BASE)
	if ds, ok := parsed.Args["ds"]; !ok {
		if ds == "china" {
			base, err = url.Parse(eb2.ESI_CHINA)
		}
	}
	if err != nil {
		// TODO: Return Parsing Error
		return Msg{}, nil
	}

	uri, err := url.ParseRequestURI(p)
	if err != nil {
		return Msg{}, nil
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
		return Msg{}, err

	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Msg{}, err
	}

	if resp.StatusCode != 200 {
		txt := "The request to %s failed with status code %d and error message %s"
		return Msg{
			Text: fmt.Sprintf(txt, uri.String(), resp.StatusCode, string(data)),
		}, nil
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

	q := url.Values{}
	q.Set("channels", parsed.ChannelID)
	q.Set("content", string(data))
	q.Set("filename", "response.json")
	q.Set("filetype", "json")
	q.Set("initial_comment", fmt.Sprintf("%s (%dms)", strings.ToUpper(resp.Status), d.Milliseconds()))
	q.Set("title", uri.String())

	uri = &url.URL{
		Scheme: "https",
		Host:   "slack.com",
		Path:   "/api/files.upload",
	}

	j := []byte(q.Encode())
	b := bytes.NewBuffer(j)

	req, err := http.NewRequest(http.MethodPost, uri.String(), b)
	if err != nil {
		return Msg{}, fmt.Errorf("failed to create request to deliver response: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.config.SlackAPIToken))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return Msg{}, fmt.Errorf("failed to execute request to slack  api to deliver file: %w", err)
	}

	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return Msg{}, fmt.Errorf("unable to read response body: %w", err)

	}

	if resp.StatusCode != 200 {
		return Msg{}, fmt.Errorf("Post Request to Slack API returned invalid status code %d", resp.StatusCode)
	}

	fmt.Println(string(data))

	return Msg{}, nil

}
