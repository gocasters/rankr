package otel

type Exporter string

const (
	ExporterGrpc    Exporter = "grpc"
	ExporterConsole Exporter = "console"
)
