package jobs

import (
	"context"
	"encoding/json"
)

type Handler func(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)

type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(jobType string, handler Handler) {
	r.handlers[jobType] = handler
}

func (r *Registry) Get(jobType string) (Handler, bool) {
	h, ok := r.handlers[jobType]
	return h, ok
}

func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}

func RegisterAll(r *Registry) {
	r.Register("image_resize", ImageResizeHandler)
	r.Register("pdf_export", PDFExportHandler)
	r.Register("data_transform", DataTransformHandler)
	r.Register("email_send", EmailSendHandler)
	r.Register("slow_compute", SlowComputeHandler)
	r.Register("always_fail", AlwaysFailHandler)
}
