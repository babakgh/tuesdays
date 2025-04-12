package tracing

import (
	"context"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
)

// Tracer interfaces for abstracting tracing implementations
type Tracer interface {
	StartSpan(name string, opts ...SpanOption) Span
	Inject(ctx context.Context, carrier interface{}) error
	Extract(carrier interface{}) (context.Context, error)
}

// Span interface for abstracting span implementations
type Span interface {
	End()
	SetAttribute(key string, value interface{})
	AddEvent(name string, attributes map[string]interface{})
	RecordError(err error)
	Context() context.Context
}

// SpanOption function for configuring span options
type SpanOption func(*SpanOptions)

// SpanOptions for span creation
type SpanOptions struct {
	Attributes map[string]interface{}
	Parent     context.Context
}

// NoopTracer is a tracer that does nothing
type NoopTracer struct{}

// StartSpan implements Tracer.StartSpan
func (t *NoopTracer) StartSpan(name string, opts ...SpanOption) Span {
	return &NoopSpan{}
}

// Inject implements Tracer.Inject
func (t *NoopTracer) Inject(ctx context.Context, carrier interface{}) error {
	return nil
}

// Extract implements Tracer.Extract
func (t *NoopTracer) Extract(carrier interface{}) (context.Context, error) {
	return context.Background(), nil
}

// NoopSpan is a span that does nothing
type NoopSpan struct{}

// End implements Span.End
func (s *NoopSpan) End() {}

// SetAttribute implements Span.SetAttribute
func (s *NoopSpan) SetAttribute(key string, value interface{}) {}

// AddEvent implements Span.AddEvent
func (s *NoopSpan) AddEvent(name string, attributes map[string]interface{}) {}

// RecordError implements Span.RecordError
func (s *NoopSpan) RecordError(err error) {}

// Context implements Span.Context
func (s *NoopSpan) Context() context.Context {
	return context.Background()
}

// WithAttributes creates a SpanOption that sets attributes on the span
func WithAttributes(attributes map[string]interface{}) SpanOption {
	return func(opts *SpanOptions) {
		if opts.Attributes == nil {
			opts.Attributes = make(map[string]interface{})
		}
		for k, v := range attributes {
			opts.Attributes[k] = v
		}
	}
}

// WithParent creates a SpanOption that sets the parent context
func WithParent(ctx context.Context) SpanOption {
	return func(opts *SpanOptions) {
		opts.Parent = ctx
	}
}

// NewTracer creates a new tracer based on the configuration
func NewTracer(cfg config.TracingConfig) (Tracer, error) {
	// Return NoopTracer if tracing is disabled
	if !cfg.Enabled {
		return &NoopTracer{}, nil
	}

	// The actual implementation will be in a subpackage
	// This provides a layer of indirection so we can swap implementations
	// We'll handle the actual initialization at a higher level
	return &NoopTracer{}, nil
}
