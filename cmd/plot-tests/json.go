package main

import (
	"encoding/json"
	"io"
	"time"
)

type Event struct {
	Action  string
	Output  string
	Time    time.Time
	Package string
	Test    string
	Elapsed float64
}

const (
	ActionRun    = "run"
	ActionOutput = "output"
	ActionPause  = "pause"
	ActionCont   = "cont"
	ActionPass   = "pass"
	ActionFail   = "fail"
	ActionSkip   = "skip"
)

type EventDecoder struct {
	json *json.Decoder
}

func NewEventDecoder(r io.Reader) *EventDecoder {
	return &EventDecoder{
		json: json.NewDecoder(r),
	}
}

func (dec *EventDecoder) Next() (Event, error) {
tryagain:
	var event Event
	err := dec.json.Decode(&event)
	if err == io.EOF {
		return Event{}, io.EOF
	}
	if err != nil {
		return Event{}, err
	}
	if event.Time.IsZero() {
		goto tryagain
	}
	if event.Action == ActionOutput {
		goto tryagain
	}

	return event, nil
}
