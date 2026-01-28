package outbox

import (
	"encoding/json"
	"time"
)

type CreateEvent struct {
	aggregateID string
	topic       string
	ttype       string
	createdAt   time.Time
	payload     json.RawMessage
}

func NewCreateEventInput() *CreateEvent {
	return &CreateEvent{}
}

func (i *CreateEvent) SetAggregateID(aggregateID string) {
	i.aggregateID = aggregateID
}

func (i *CreateEvent) SetTopic(t string) {
	i.topic = t
}

func (i *CreateEvent) SetType(t string) {
	i.ttype = t
}

func (i *CreateEvent) SetPayload(p json.RawMessage) {
	i.payload = p
}

func (i *CreateEvent) SetCreatedAt(t time.Time) {
	i.createdAt = t
}

func (i *CreateEvent) GetAggregateID() string {
	return i.aggregateID
}

func (i *CreateEvent) GetTopic() string {
	return i.topic
}

func (i *CreateEvent) GetType() string {
	return i.ttype
}

func (i *CreateEvent) GetPayload() json.RawMessage {
	return i.payload
}

func (i *CreateEvent) GetCreatedAt() time.Time {
	return i.createdAt
}
