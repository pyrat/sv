package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	IconSuccess = "🟢"
	IconWarn    = "🟡"
	IconError   = "🔴"
)

type Client struct {
	webhookURL string
}

func slackNew(webhookURL string) *Client {
	return &Client{webhookURL: webhookURL}
}

type slackRequestBody struct {
	Text      string `json:"text"`
	Channel   string `json:"channel"`
	Username  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
}

// SendNotification sends notification to specified channel with username
func (c *Client) SendNotification(channel, username, msg string) error {
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	return c.sendSlackNotification(slackRequestBody{
		Text:     msg,
		Channel:  channel,
		Username: username,
	})
}

// SendAlert sends critical alert notification to #biomass-frame-alerts channel
func (c *Client) SendAlert(icon, msg string) error {
	return c.sendSlackNotification(slackRequestBody{
		Text:      fmt.Sprintf("%s %s", icon, msg),
		Channel:   "#smartwinch",
		Username:  "wiggle",
		IconEmoji: "robot_face",
	})
}

func (c *Client) sendSlackNotification(slackReqBody slackRequestBody) error {
	if c.webhookURL == "" {
		log.Warn().Msg(slackReqBody.Text)
		return nil
	}

	slackBody, err := json.Marshal(slackReqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 15 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint: gosec

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	if buf.String() != "ok" {
		return errors.New("non-ok response returned from Slack")
	}
	return nil
}
