package goslack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Slack struct {
	webHook                string
	authorLink, authorName string
	appName, emoji         string
}

//New slack client
func New(hook, authorLink, authorName, appName, emoji string) *Slack {
	return &Slack{webHook: hook, authorLink: authorLink, authorName: authorName, appName: appName, emoji: emoji}

}

func (s *Slack) NewMessage() *Message {
	return &Message{service: s}

}

//Message A slack message
type Message struct {
	subject, color, msg string
	ts                  int64
	fields              []Field
	service             *Slack
}

func (m *Message) Color(color string) *Message {
	if len(color) == 7 && color[0] == '#' {
		m.color = color
	} else {
		m.color = `warning`
	}
	return m
}
func (m *Message) Subject(subject string) *Message {
	m.subject = subject
	return m
}
func (m *Message) Text(text string) *Message {
	m.msg = text
	return m
}

func (m *Message) Fields(fields []Field) *Message {
	m.fields = fields
	return m
}
func (m *Message) TS(tm time.Time) *Message {
	m.ts = tm.Unix()
	return m
}

var colors = []string{`good`, `warning`, `danger`}

const (
	Disappointed = `disappointed`
)

//Attachment []Attachment represent attachments in slack body
type Attachment struct {
	Fallback   string  `json:"fallback"`
	Text       string  `json:"text"`
	Color      string  `json:"color,omitempty"`
	AuthorLink string  `json:"author_link,omitempty"`
	AuthorName string  `json:"author_name,omitempty"`
	Title      string  `json:"title,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
	ImageURL   string  `json:"image_url,omitempty"`
	ThumbURL   string  `json:"thumb_url,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	FooterIcon string  `json:"footer_icon,omitempty"`
	TS         int64   `json:"ts,omitempty"`
	Pretext    string  `json:"pretext,omitempty"`
	AuthorIcon string  `json:"author_icon,omitempty"`
	TitleLink  string  `json:"title_link,omitempty"`
}

//Field []Field represent fields in slack body
type Field struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

//Slack API body
type body struct {
	AppName     string       `json:"username,omitempty"`
	Emoji       string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments"`
}

//Send send a message synchronously.
func (m *Message) Send() error {
	//fillMessageDefaults(m)
	bd := body{
		Attachments: []Attachment{
			Attachment{
				Fallback:   m.subject,
				Text:       m.msg,
				Color:      m.color,
				AuthorLink: m.service.authorLink,
				AuthorName: m.service.authorName,
				Title:      m.subject,
				Fields:     m.fields,
			},
		},
	}
	if len(m.service.emoji) > 0 {
		bd.Emoji = m.service.emoji
	}
	if len(m.service.appName) > 0 {
		bd.AppName = m.service.appName
	}
	b, err := json.Marshal(bd)
	log.Println(string(b))
	if err != nil {
		return err
	}
	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, m.service.webHook, bytes.NewBuffer(b)); err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := clientWithTimeout().Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		o := &bytes.Buffer{}
		io.Copy(o, resp.Body)
		return fmt.Errorf("code: %d, message:%s", resp.StatusCode, o.String())
	}
	return nil
}

func fillMessageDefaults(msg *Message) {
	if len(msg.color) == 0 {
		msg.color = `warning`
	}
	if len(msg.msg) == 0 {
		msg.color = `no text`
	}
	if len(msg.subject) == 0 {
		msg.color = `no subject`
	}
	if msg.ts == 0 {
		msg.ts = time.Now().UTC().Unix()
	}

}

func clientWithTimeout() *http.Client {
	return &http.Client{
		Timeout: time.Second * 30,
	}
}
