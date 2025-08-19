package otel

import (
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

// defaultResourceBuilder creates standard OpenTelemetry resources
type defaultResourceBuilder struct{}

func newDefaultResourceBuilder() ResourceBuilder {
	return &defaultResourceBuilder{}
}

func (r *defaultResourceBuilder) BuildResource(serviceName string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)
}
