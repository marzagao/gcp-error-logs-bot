package bot

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Bot struct {
	SlackWebhookURL string
	LogEntrySource  IncidentLogEntrySource
	Log             *logrus.Logger
}

func NewBot(log *logrus.Logger, projectID string, logName string, timezone string, webhookURL string) (*Bot, error) {
	return &Bot{
		SlackWebhookURL: webhookURL,
		LogEntrySource: &GCPLogAdmin{
			ProjectID: projectID,
			LogName:   logName,
			Timezone:  timezone,
		},
		Log: log,
	}, nil
}

func (b *Bot) Run(input string) {
	botOutput := []string{}
	logEntries := []IncidentLogEntry{}
	payload := IncidentPayload{}
	err := payload.Populate(input)
	if err != nil {
		botOutput = append(botOutput, fmt.Sprintf("Failed to parse input: %v\n", err))
		botOutput = append(botOutput, input)
		b.postToWebhook(botOutput, logEntries, "")
		return
	}
	incidentURL := payload.Incident.URL
	incidentURLParts := strings.Split(incidentURL, "/")
	incidentID := incidentURLParts[len(incidentURLParts)-1]
	botOutput = append(botOutput, fmt.Sprintf("Incident <%s|%s>\n", incidentURL, incidentID))
	b.Log.Debugf("getting incidents using filter: %s", b.LogEntrySource.GetLogAdminFilter(payload))
	logEntries, err = b.LogEntrySource.GetLogEntries(payload)
	b.Log.Debugf("found %d log entries matching above filter", len(logEntries))
	if err != nil {
		botOutput = append(botOutput, fmt.Sprintf("Failed to retrieve log entries: %v\n", err))
		b.postToWebhook(botOutput, logEntries, payload.Incident.State)
		return
	}
	b.postToWebhook(botOutput, logEntries, payload.Incident.State)
}
