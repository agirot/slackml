package slack

import (
	"context"
	"fmt"
	"github.com/agirot/slackml/internal/helper"
	"github.com/parnurzeal/gorequest"
)

type sendMessage struct {
	Channel string   `json:"channel"`
	Blocks  []blocks `json:"blocks"`
}

type blocks struct {
	Type           string     `json:"type,omitempty"`
	Text           *text      `json:"text,omitempty"`
	SlackAccessory *accessory `json:"accessory,omitempty"`
}

type text struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type accessory struct {
	Type     string `json:"type,omitempty"`
	Text     text   `json:"text,omitempty"`
	Value    string `json:"value,omitempty"`
	URL      string `json:"url,omitempty"`
	ActionID string `json:"action_id,omitempty"`
}

func SendSlack(ctx context.Context, dateStr, from, title, url string) error {
	splitBlocks := make([]blocks, 0, 2)
	header := blocks{
		Type: "section",
		Text: &text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("`%s` _%s_ ", from, dateStr),
		}}

	body := blocks{
		Type: "section",
		Text: &text{
			Type: "mrkdwn",
			Text: ">" + title,
		},
		SlackAccessory: &accessory{
			Type: "button",
			Text: text{
				Type: "plain_text",
				Text: "Link",
			},
			Value:    "click_me_123",
			URL:      url,
			ActionID: "button-action",
		},
	}

	splitBlocks = append(splitBlocks, header, body)

	cfg := helper.GetConfigContext(ctx)
	_, _, errs := gorequest.New().Post("https://slack.com/api/chat.postMessage?channel="+cfg.SlackChannelID).
		Set("Content-type", "application/json").
		Set("Authorization", fmt.Sprintf("Bearer %v", cfg.SlackToken)).
		SendStruct(sendMessage{Channel: cfg.SlackChannelID, Blocks: splitBlocks}).
		End()

	if len(errs) > 0 {
		return fmt.Errorf("failed to POST slack: %w", errs[0])
	}
	return nil
}
