package prof

func init() {
	GlobalProfiler = &NOPProfiler{}
}

type (
	Profiler interface {
		PublishSpan(span Span)
	}

	NOPProfiler struct{}
)

var (
	GlobalProfiler Profiler
)

func (p *NOPProfiler) PublishSpan(span Span) {}
