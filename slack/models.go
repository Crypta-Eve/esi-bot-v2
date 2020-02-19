package slack

import (
	"net/url"
	"strings"
)

// The following were copied and modified appropriately
// to serve this applications needs from https://github.com/nlopes/slack

// Attachment contains all the information for an attachment
type Attachment struct {
	Title     string             `json:"title,omitempty"`
	TitleLink string             `json:"title_link,omitempty"`
	Pretext   string             `json:"pretext,omitempty"`
	Color     string             `json:"color,omitempty"`
	Text      string             `json:"text,omitempty"`
	Actions   []AttachmentAction `json:"actions,omitempty"`
	Footer    string             `json:"footer,omitempty"`
	Fallback  string             `json:"fallback,omitempty"`
	Fields    []AttachmentField  `json:"fields,omitempty"`
}

// AttachmentAction is a button or menu to be included in the attachment. Required when
// using message buttons or menus and otherwise not useful. A maximum of 5 actions may be
// provided per attachment.
type AttachmentAction struct {
	Name  string `json:"name"`            // Required.
	Text  string `json:"text"`            // Required.
	Style string `json:"style,omitempty"` // Optional. Allowed values: "default", "primary", "danger".
	Type  string `json:"type"`            // Required. Must be set to "button" or "select".
	Value string `json:"value,omitempty"` // Optional.
	URL   string `json:"url,omitempty"`   // Optional.
}

// AttachmentField contains information for an attachment field
// An Attachment can contain multiple of these
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Msg contains information about a slack message
type Msg struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
	// // file_share, file_comment, file_mention
	// Files []File `json:"files,omitempty"`

	// slash commands and interactive messages
	ResponseType    string `json:"response_type,omitempty"`
	ReplaceOriginal bool   `json:"replace_original"`
	DeleteOriginal  bool   `json:"delete_original"`
	UnfurlLinks     bool   `json:"unfurl_links"`
}

// SlashCommand contains information about a request of the slash command
type SlashCommand struct {
	Token          string `json:"token"`
	TeamID         string `json:"team_id"`
	TeamDomain     string `json:"team_domain"`
	EnterpriseID   string `json:"enterprise_id,omitempty"`
	EnterpriseName string `json:"enterprise_name,omitempty"`
	ChannelID      string `json:"channel_id"`
	ChannelName    string `json:"channel_name"`
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	Command        string `json:"command"`
	Text           string `json:"text"`
	Args           map[string]string
	ResponseURL    string `json:"response_url"`
	TriggerID      string `json:"trigger_id"`
}

// ParseSlashCommand will parse the request of the slash command
func ParseSlashCommand(values url.Values) (s SlashCommand, err error) {

	s.Token = values.Get("token")
	s.TeamID = values.Get("team_id")
	s.TeamDomain = values.Get("team_domain")
	s.EnterpriseID = values.Get("enterprise_id")
	s.EnterpriseName = values.Get("enterprise_name")
	s.ChannelID = values.Get("channel_id")
	s.ChannelName = values.Get("channel_name")
	s.UserID = values.Get("user_id")
	s.UserName = values.Get("user_name")
	s.Command = values.Get("command")
	text := values.Get("text")
	args := strings.Split(text, " ")
	if len(args) > 1 {
		s.Args = make(map[string]string, 0)
		for _, arg := range args[1:] {
			argAttr := strings.Split(arg, "=")
			if len(argAttr) > 2 {
				continue
			}
			key := argAttr[0]
			value := argAttr[1]
			if !strings.HasPrefix(key, "--") {
				continue
			}
			key = strings.Join(strings.Split(key, "--")[1:], "")
			s.Args[key] = value
		}
	}
	s.Text = args[0]

	s.ResponseURL = values.Get("response_url")
	s.TriggerID = values.Get("trigger_id")
	return s, nil
}
