package main

import "time"

type Span struct {
	Start  time.Time
	Finish time.Time
}

func (span *Span) Duration() time.Duration {
	return span.Finish.Sub(span.Start)
}

func (span *Span) Extend(t time.Time) {
	if span.Start.IsZero() {
		span.Start = t
	}
	if span.Finish.IsZero() {
		span.Finish = t
	}

	if t.Before(span.Start) {
		span.Start = t
	}
	if t.After(span.Finish) {
		span.Finish = t
	}
}
