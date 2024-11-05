package discord

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
	"encore.dev/pubsub"
	"encore.app/monitor"
)

type NotifyParams struct {
    // Content is the Discord message text to send.
    Content string `json:"content"`
}

// Notify sends a Discord message to a pre-configured channel using a
// Discord Webhook (see https://discord.com/developers/docs/resources/webhook).
//
//encore:api private
func Notify(ctx context.Context, p *NotifyParams) error {
    reqBody, err := json.Marshal(p)
    if err != nil {
        return err
    }
    req, err := http.NewRequestWithContext(ctx, "POST", secrets.DiscordWebhookURL, bytes.NewReader(reqBody))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("notify discord: %s: %s", resp.Status, body)
    }
    return nil
}

var _ = pubsub.NewSubscription(monitor.TransitionTopic, "slack-notification", pubsub.SubscriptionConfig[*monitor.TransitionEvent]{
	Handler: func(ctx context.Context, event *monitor.TransitionEvent) error {
		// Compose message
		msg := fmt.Sprintf("😨 *%s is down!*", event.Site.URL)
		if event.Up {
			msg = fmt.Sprintf("😄 *%s is back up.*", event.Site.URL)
		}

		// Send the Discord msg
		return Notify(ctx, &NotifyParams{Content: msg})
	},
})

var secrets struct {
    // DiscordWebhookURL defines the Discord webhook URL to send
    // uptime notifications to.
    DiscordWebhookURL string
}