package bot

import (
	"encoding/json"
	"time"
)

const precisionAllowanceInSeconds = 90

type IncidentPayload struct {
	Incident Incident
	Version  string
}

type Incident struct {
	IncidentID    string `json:"incident_id"`
	ResourceID    string `json:"resource_id"`
	ResourceName  string `json:"resource_name"`
	State         string `json:"state"`
	StartedAt     int64  `json:"started_at"`
	EndedAt       *int64 `json:"ended_at"`
	PolicyName    string `json:"policy_name"`
	ConditionName string `json:"condition_name"`
	URL           string `json:"url"`
	Summary       string `json:"summary"`
}

type IncidentLogEntrySource interface {
	GetLogAdminFilter(payload IncidentPayload) string
	GetLogEntries(payload IncidentPayload) ([]IncidentLogEntry, error)
}

type IncidentLogEntry struct {
	EventTime    time.Time
	Environment  string
	ResourceID   string
	ResourceName string
	Description  string
	Error        string
	Caller       string
	Service      string
}

func (i *Incident) StartedAtTime() string {
	return time.Unix(i.StartedAt-precisionAllowanceInSeconds, 0).UTC().Format(time.RFC3339)
}

func (i *Incident) EndedAtTime() string {
	if i.EndedAt != nil {
		return time.Unix(*i.EndedAt, 0).UTC().Format(time.RFC3339)
	}
	return time.Now().UTC().Format(time.RFC3339)
}

func (i *IncidentPayload) Populate(input string) error {
	return json.Unmarshal([]byte(input), i)
}
