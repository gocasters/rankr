package database

type Config struct {
	Host              string `koanf:"host"`
	Port              int    `koanf:"port"`
	Username          string `koanf:"user"`
	Password          string `koanf:"password"`
	DBName            string `koanf:"db_name"`
	SSLMode           string `koanf:"ssl_mode"`
	MaxConns          int32  `koanf:"max_conns"`
	MinConns          int32  `koanf:"min_conns"`
	MaxConnLifetime   int    `koanf:"max_conn_lifetime"`
	MaxConnIdleTime   int    `koanf:"max_conn_idle_time"`
	HealthCheckPeriod int    `koanf:"health_check_period"`
	PathOfMigrations  string `koanf:"path_of_migrations"`
}
