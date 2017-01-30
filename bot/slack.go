package bot

import (
	"strings"

	"github.com/bluele/slack"
)

func (b *Bot) postToWebhook(output []string, logEntries []IncidentLogEntry, incidentState string) {
	outputString := strings.Join(output, "\n")
	instanceIDs := []string{}
	for _, logEntry := range logEntries {
		instanceIDs = append(instanceIDs, logEntry.ResourceID)
	}
	instanceIPs, err := b.getInstanceIPs(instanceIDs)
	if err != nil {
		b.Log.Errorf("Failed to retrieve instance IPs: %s", err)
	}
	attachments := []*slack.Attachment{}
	for _, logEntry := range logEntries {
		color := "good"
		if incidentState == "open" {
			color = "warning"
			if logEntry.Environment == "prd" {
				color = "danger"
			}
		}
		if logEntry.Caller != "" {
			logEntry.Caller = "```" + logEntry.Caller + "```"
		}
		attachments = append(attachments, &slack.Attachment{
			Fallback:   logEntry.Description + " : " + logEntry.Error,
			Color:      color,
			MarkdownIn: []string{"fields"},
			Fields: []*slack.AttachmentField{
				{
					Title: "Message",
					Value: "```" + logEntry.Description + " : " + logEntry.Error + "```",
					Short: false,
				},
				{
					Title: "Caller",
					Value: logEntry.Caller,
					Short: false,
				},
				{
					Title: "Environment",
					Value: logEntry.Environment,
					Short: true,
				},
				{
					Title: "Event time",
					Value: logEntry.EventTime.String(),
					Short: true,
				},
				{
					Title: "Service",
					Value: logEntry.Service,
					Short: true,
				},
				{
					Title: "Instance IP",
					Value: instanceIPs[logEntry.ResourceID],
					Short: true,
				},
			},
		})
	}
	webhook := slack.NewWebHook(b.SlackWebhookURL)
	webhookPayload := &slack.WebHookPostPayload{
		Text:        outputString,
		Attachments: attachments,
	}
	err = webhook.PostMessage(webhookPayload)
	if err != nil {
		b.Log.Fatalf("Failed to post to Slack webhook: %v\n", err)
		return
	}
}
