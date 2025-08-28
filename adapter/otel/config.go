package otel

type Config struct {
	Endpoint    string   `koanf:"endpoint"`
	ServiceName string   `koanf:"service_name"`
	Exporter    Exporter `koanf:"exporter"`
}
