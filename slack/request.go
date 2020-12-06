package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eveisesi/eb2"
	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

// func (s *service) makeESITypeRequestMessage(event Event) {

// 	if len(event.args) != 1 {
// 		err := errors.New("this command excepts a single argument, which should be the id of the item you are need details for")
// 		// TODO: Return Parsing Error
// 		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), false))
// 	}

// 	event.trigger = fmt.Sprintf("/latest/universe/types/%s", event.args[0])
// 	s.makeESIDynamicRequestMessage(event)
// }

// The below types are taken from the goesi library by antihax
type GetUniverseTypesTypeIdOk struct {
	Capacity        float32                                 `json:"capacity,omitempty"`         /* capacity number */
	Description     string                                  `json:"description,omitempty"`      /* description string */
	DogmaAttributes []*GetUniverseTypesTypeIdDogmaAttribute `json:"dogma_attributes,omitempty"` /* dogma_attributes array */
	DogmaEffects    []*GetUniverseTypesTypeIdDogmaEffect    `json:"dogma_effects,omitempty"`    /* dogma_effects array */
	GraphicId       int32                                   `json:"graphic_id,omitempty"`       /* graphic_id integer */
	GroupId         int32                                   `json:"group_id,omitempty"`         /* group_id integer */
	IconId          int32                                   `json:"icon_id,omitempty"`          /* icon_id integer */
	MarketGroupId   int32                                   `json:"market_group_id,omitempty"`  /* This only exists for types that can be put on the market */
	Mass            float32                                 `json:"mass,omitempty"`             /* mass number */
	Name            string                                  `json:"name,omitempty"`             /* name string */
	PackagedVolume  float32                                 `json:"packaged_volume,omitempty"`  /* packaged_volume number */
	PortionSize     int32                                   `json:"portion_size,omitempty"`     /* portion_size integer */
	Published       bool                                    `json:"published,omitempty"`        /* published boolean */
	Radius          float32                                 `json:"radius,omitempty"`           /* radius number */
	TypeId          int32                                   `json:"type_id,omitempty"`          /* type_id integer */
	Volume          float32                                 `json:"volume,omitempty"`           /* volume number */
}

type GetUniverseTypesTypeIdDogmaAttribute struct {
	AttributeId int32   `json:"attribute_id,omitempty"` /* attribute_id integer */
	Name        string  `json:"name,omitempty"`
	Value       float32 `json:"value,omitempty"` /* value number */
}

type GetUniverseTypesTypeIdDogmaEffect struct {
	EffectId  int32  `json:"effect_id,omitempty"` /* effect_id integer */
	Name      string `json:"name,omitempty"`
	IsDefault bool   `json:"is_default,omitempty"` /* is_default boolean */
}

type GetDogmaAttributesAttributeIdOk struct {
	Name string `json:"name,omitempty"` /* name string */
}

type GetDogmaEffectsEffectIdOk struct {
	Name string `json:"name,omitempty"` /* name string */
}

var wg sync.WaitGroup

func (s *service) makeESITypeRequestMessage(event Event) {
	if len(event.args) == 0 {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText("You need to supply at least one type id to lookup", false))
		return
	} else if len(event.args) > 10 {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText("Please supply a maximum of 10 ids to look up", false))
		return
	}

	for _, a := range event.args {
		start := time.Now()

		id, err := strconv.Atoi(a)
		if err != nil {
			_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(fmt.Sprintf("%s is not a valid number, skipping", a), false))
			continue
		}

		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			var internalWg sync.WaitGroup
			esiCalls := 0

			url := fmt.Sprintf("https://esi.evetech.net/latest/universe/types/%d/", id)

			esiCalls++
			resp, err := s.client.Get(url)
			if err != nil {
				_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(fmt.Sprintf("Failed to check id %d. Error: %s", id, err.Error()), false))
				return
			}

			bd, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(fmt.Sprintf("Failed to decode response from esi for id %d. Error: %s", id, err.Error()), false))
				return
			}

			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(fmt.Sprintf("Got an error from esi. Code %d, Body: %s", id, string(bd)), false))
				return
			}

			var tp = &GetUniverseTypesTypeIdOk{}
			err = json.Unmarshal(bd, tp)
			if err != nil {
				_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(fmt.Sprintf("Failed to decode response from esi for id %d. Error: %s", id, err.Error()), false))
				return
			}

			// Now that we have the type, iterate over all dogma attributes to resolve them (if present)
			for _, dgm := range tp.DogmaAttributes {
				internalWg.Add(1)
				go func(dgm *GetUniverseTypesTypeIdDogmaAttribute) {
					defer internalWg.Done()
					url = fmt.Sprintf("https://esi.evetech.net/v1/dogma/attributes/%d/", dgm.AttributeId)

					esiCalls++
					res, err := s.client.Get(url)
					if err != nil {
						return
					}

					bdd, err := ioutil.ReadAll(res.Body)
					if err != nil {
						return
					}

					resp.Body.Close()

					if res.StatusCode != http.StatusOK {
						dgm.Name = "failed to acquire the name of this attribute"
					}

					var attr GetDogmaAttributesAttributeIdOk
					err = json.Unmarshal(bdd, &attr)
					if err != nil {
						return
					}

					dgm.Name = attr.Name

				}(dgm)
			}

			// Now that we have the type, iterate over all dogma effects to resolve them (if present)
			for _, dgm := range tp.DogmaEffects {
				internalWg.Add(1)
				go func(dgm *GetUniverseTypesTypeIdDogmaEffect) {
					defer internalWg.Done()
					url = fmt.Sprintf("https://esi.evetech.net/v1/dogma/effects/%d/", dgm.EffectId)

					esiCalls++
					res, err := s.client.Get(url)
					if err != nil {
						return
					}

					bdd, err := ioutil.ReadAll(res.Body)
					if err != nil {
						return
					}

					resp.Body.Close()

					if res.StatusCode != http.StatusOK {
						dgm.Name = "failed to acquire the name of this attribute"
					}

					var attr GetDogmaAttributesAttributeIdOk
					err = json.Unmarshal(bdd, &attr)
					if err != nil {
						return
					}

					dgm.Name = attr.Name
				}(dgm)
			}
			internalWg.Wait()

			data, err := json.Marshal(tp)
			if err != nil {
				_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText("Internal Error Attempting to Marshal Response", false))
				return
			}
			// From this point down I am just copying from DDs code.. Shameless rip

			dst := new(bytes.Buffer)
			_ = json.Indent(dst, data, "", "   ")

			data = dst.Bytes()

			if len(data) > 1024000 {
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
				InitialComment: fmt.Sprintf("%s (%dms) (%d calls to esi)", strings.ToUpper(resp.Status), d.Milliseconds(), esiCalls),
				Title:          tp.Name,
			})
			if err != nil {
				s.logger.WithError(err).Error("failed to respond to request for esi data.")
				return
			}
			s.logger.Info("successfully responded to request for esi data")
		}(id)
	}
	wg.Wait()

}

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
	base, err = url.Parse(eb2.ESI_URLS[eb2.ESI_TRANQUILITY])
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

	if len(data) > 1024000 {
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
					// We are at the end of the route and everything has matched,
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
