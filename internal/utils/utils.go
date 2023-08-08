package utils

import (
	"github.com/SmoothWay/discord-bot/internal/config"
	"github.com/getsentry/sentry-go"
)

func SendChannelMessage(channelID string, message string) {
	_, err := config.Dg.ChannelMessageSend(channelID, message)
	if err != nil {
		sentry.CaptureException(err)
	}
}
