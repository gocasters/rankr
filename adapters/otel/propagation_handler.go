package otel

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

// otelPropagationHandler handles context propagation
type otelPropagationHandler struct {
	propagator propagation.TextMapPropagator
}

func newOtelPropagationHandler() PropagationHandler {
	return &otelPropagationHandler{
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}
}

func (p *otelPropagationHandler) GetCarrierFromContext(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	p.propagator.Inject(ctx, carrier)
	return carrier
}

func (p *otelPropagationHandler) GetContextFromCarrier(carrier map[string]string) context.Context {
	c := propagation.MapCarrier{}
	for k, v := range carrier {
		c.Set(k, v)
	}
	return p.propagator.Extract(context.Background(), c)
}
