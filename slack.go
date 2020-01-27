package cachet

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type Slack struct {
	WebhookURL  string
	Attachments []Attachments `json:"attachments"`
}
type Fields struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
type Attachments struct {
	Fallback   string   `json:"fallback"`
	Color      string   `json:"color"`
	Pretext    string   `json:"pretext"`
	Title      string   `json:"title"`
	TitleLink  string   `json:"title_link"`
	Text       string   `json:"text"`
	Fields     []Fields `json:"fields"`
	ThumbURL   string   `json:"thumb_url"`
	Footer     string   `json:"footer"`
	FooterIcon string   `json:"footer_icon"`
	Ts         int64    `json:"ts"`
}

func test() {
	slack := Slack{
		Attachments: []Attachments{
			Attachments{
				Fallback:   "Required plain-text summary of the attachment.",
				Color:      "#36a64f",
				Title:      "Slack API Documentation",
				TitleLink:  "https://status.easyship.com",
				Text:       "Optional text that appears within the attachment",
				Footer:     "Cachet Monitor",
				FooterIcon: "https://i.imgur.com/spck1w6.png",
				Ts:         time.Now().Unix(),
			},
		}}
	slack.WebhookURL = "https://hooks.slack.com/services/0000000/00000000/xxxxxxxxxxxxxxxxxxx"
	err := slack.SendSlackNotification()
	if err != nil {
		log.Fatal(err)
	}
}

// SendSlackNotification will post to an 'Incoming Webook' url setup in Slack Apps. It accepts
// some text and the slack channel is saved within Slack.
func (slack *Slack) SendSlackNotification() error {

	slackBody, _ := json.Marshal(slack)
	req, err := http.NewRequest(http.MethodPost, slack.WebhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}
