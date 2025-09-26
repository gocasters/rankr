package database

// Config represents the configuration for a database connection.
type Config struct {
	Host              string `koanf:"host"`
	Port              int    `koanf:"port"`
	Username          string `koanf:"user"`
	Password          string `koanf:"password"`
	DBName            string `koanf:"db_name"`
	Schema            string `koanf:"schema"`
	SSLMode           string `koanf:"ssl_mode"`
	MaxConns          int32  `koanf:"max_conns"`
	MinConns          int32  `koanf:"min_conns"`
	MaxConnLifetime   int    `koanf:"max_conn_lifetime"`
	MaxConnIdleTime   int    `koanf:"max_conn_idle_time"`
	HealthCheckPeriod int    `koanf:"health_check_period"`
	PathOfMigrations  string `koanf:"path_of_migrations"`
}

// ConfigOption defines a function that modifies a Config.
type ConfigOption func(*Config)

// NewConfig creates a new Config with default values and applies the provided options.
func NewConfig(opts ...ConfigOption) Config {
	cfg := Config{
		Host:              "127.0.0.1",
		Port:              5432,
		Username:          "root",
		Password:          "",
		DBName:            "dbname",
		Schema:            "public",
		SSLMode:           "disable",
		MaxConns:          10,   // Default value
		MinConns:          2,    // Default value
		MaxConnLifetime:   3600, // Default: 1 hour in seconds
		MaxConnIdleTime:   600,  // Default: 10 minutes in seconds
		HealthCheckPeriod: 60,   // Default: 1 minute in seconds
		PathOfMigrations:  "./migrations",
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

// WithHost sets the Host field of the Config.
func WithHost(host string) ConfigOption {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort sets the Port field of the Config.
func WithPort(port int) ConfigOption {
	return func(c *Config) {
		c.Port = port
	}
}

// WithUsername sets the Username field of the Config.
func WithUsername(username string) ConfigOption {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword sets the Password field of the Config.
func WithPassword(password string) ConfigOption {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDBName sets the DBName field of the Config.
func WithDBName(dbName string) ConfigOption {
	return func(c *Config) {
		c.DBName = dbName
	}
}

// WithSSLMode sets the SSLMode field of the Config.
func WithSSLMode(sslMode string) ConfigOption {
	return func(c *Config) {
		c.SSLMode = sslMode
	}
}

// WithMaxConns sets the MaxConns field of the Config.
func WithMaxConns(maxConns int32) ConfigOption {
	return func(c *Config) {
		c.MaxConns = maxConns
	}
}

// WithMinConns sets the MinConns field of the Config.
func WithMinConns(minConns int32) ConfigOption {
	return func(c *Config) {
		c.MinConns = minConns
	}
}

// WithMaxConnLifetime sets the MaxConnLifetime field of the Config.
func WithMaxConnLifetime(maxConnLifetime int) ConfigOption {
	return func(c *Config) {
		c.MaxConnLifetime = maxConnLifetime
	}
}

// WithMaxConnIdleTime sets the MaxConnIdleTime field of the Config.
func WithMaxConnIdleTime(maxConnIdleTime int) ConfigOption {
	return func(c *Config) {
		c.MaxConnIdleTime = maxConnIdleTime
	}
}

// WithHealthCheckPeriod sets the HealthCheckPeriod field of the Config.
func WithHealthCheckPeriod(healthCheckPeriod int) ConfigOption {
	return func(c *Config) {
		c.HealthCheckPeriod = healthCheckPeriod
	}
}

// WithPathOfMigrations sets the PathOfMigrations field of the Config.
func WithPathOfMigrations(pathOfMigrations string) ConfigOption {
	return func(c *Config) {
		c.PathOfMigrations = pathOfMigrations
	}
}
