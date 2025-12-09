package outbox

import (
	"encoding/json"
)

type successEvent struct {
	ID   string
	Type string
}

func (se *successEvent) GetID() string {
	return se.ID
}

func (se *successEvent) GetType() string {
	return se.Type
}

type failedEvent struct {
	ID    string
	Type  string
	Error error
}

func (fe *failedEvent) GetID() string {
	return fe.ID
}

func (fe *failedEvent) GetType() string {
	return fe.Type
}

func (fe *failedEvent) GetError() error {
	return fe.Error
}

type messageMetadata struct {
	ID   string
	Type string
}

type FailedEvent interface {
	GetID() string
	GetType() string
	GetError() error
}

type SuccessEvent interface {
	GetID() string
	GetType() string
}

// type EventStatus string

// func (es EventStatus) String() string {
// 	return string(es)
// }

type Event interface {
	GetID() string
	GetAggregateID() string
	GetTopic() string
	GetType() string
	GetPayload() json.RawMessage
}

// const EventStatusNew EventStatus = "new"
// const EventStatusDone EventStatus = "done"

// type Event struct {
// 	ID          string
// 	AggregateID string
// 	Topic       string
// 	Type        string
// 	Payload     json.RawMessage
// 	Status      EventStatus
// 	CreatedAt   time.Time
// 	ReservedTo  *time.Time
// }
