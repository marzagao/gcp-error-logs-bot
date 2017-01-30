package bot

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/api/iterator"
)

type GCPLogAdmin struct {
	ProjectID string
	LogName   string
	Timezone  string
}

func (g *GCPLogAdmin) GetLogAdminFilter(payload IncidentPayload) string {
	return fmt.Sprintf(`logName = "projects/%s/logs/%s" AND timestamp >= %q AND timestamp <= %q`,
		g.ProjectID,
		g.LogName,
		payload.Incident.StartedAtTime(),
		payload.Incident.EndedAtTime())
}

func (g *GCPLogAdmin) GetLogEntries(payload IncidentPayload) ([]IncidentLogEntry, error) {
	incidentEntries := []IncidentLogEntry{}
	ctx := context.Background()
	adminClient, err := logadmin.NewClient(ctx, g.ProjectID)
	if err != nil {
		return incidentEntries, fmt.Errorf("Failed to create logadmin client: %v\n", err)
	}
	var logEntries []*logging.Entry
	iter := adminClient.Entries(ctx,
		logadmin.Filter(g.GetLogAdminFilter(payload)),
		logadmin.NewestFirst(),
	)
	retrievalCount := 0
	for {
		entry, err := iter.Next()
		retrievalCount = retrievalCount + 1
		if err == iterator.Done {
			break
		}
		if retrievalCount >= 50 {
			break
		}
		if err != nil {
			fmt.Printf("Failed to get next log entry: %v\n", err)
			continue
		}
		logEntries = append(logEntries, entry)
	}
	for _, entry := range logEntries {
		if entry == nil {
			continue
		}
		entryPayload := entry.Payload
		if value, ok := entryPayload.(*structpb.Struct); ok {
			if value == nil {
				continue
			}
			newIncidentEntry := IncidentLogEntry{}
			newIncidentEntry.Environment = payloadFieldToString(*value, "environment")
			newIncidentEntry.Description = payloadFieldToString(*value, "message")
			newIncidentEntry.Error = payloadFieldToString(*value, "error")
			newIncidentEntry.Caller = payloadFieldToString(*value, "caller")
			eventTimeString := payloadFieldToString(*value, "eventTime")
			eventTime, err := timeStringToZone(eventTimeString, g.Timezone)
			if err != nil {
				fmt.Printf("Failed to parse date %s from log entry: %v\n", eventTimeString, err)
				continue
			}
			newIncidentEntry.EventTime = eventTime
			serviceContext := value.Fields["serviceContext"]
			if serviceContext != nil {
				contextStruct := *serviceContext.GetStructValue()
				newIncidentEntry.Service = contextStruct.Fields["service"].GetStringValue()
			}
			verboseResourceName := payload.Incident.ResourceName
			if strings.Contains(verboseResourceName, "Amazon EC2 Instance labels") {
				r, _ := regexp.Compile(`instance_id=(\S*)`)
				result := r.FindStringSubmatch(verboseResourceName)
				if len(result) >= 2 {
					newIncidentEntry.ResourceID = result[1]
				}
			}
			incidentEntries = append(incidentEntries, newIncidentEntry)
		}
	}
	return incidentEntries, nil
}

func payloadFieldToString(payload structpb.Struct, fieldName string) string {
	stringValue := ""
	if payload.Fields[fieldName] != nil {
		stringValue = (*payload.Fields[fieldName]).GetStringValue()
	}
	return stringValue
}

func timeStringToZone(timeString string, locationString string) (time.Time, error) {
	location, err := time.LoadLocation(locationString)
	if err != nil {
		return time.Now(), fmt.Errorf("Failed to parse timezone location for log entry: %v\n", err)
	}
	eventTime, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		fmt.Printf("Failed to parse time from log entry: %v\n", err)
		return time.Now(), err
	}
	return eventTime.In(location), nil
}
