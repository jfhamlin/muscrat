package prof

import (
	"context"
	"time"
)

type (
	Span struct {
		name     string
		finished bool
		start    time.Time
		duration time.Duration
	}
)

func StartSpan(ctx context.Context, name string) Span {
	return Span{
		start: time.Now(),
		name:  name,
	}
}

func (s Span) Finish() {
	if s.finished {
		return
	}
	s.finished = true

	s.duration = time.Since(s.start)

	GlobalProfiler.PublishSpan(s)
}
