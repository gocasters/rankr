package database

import "fmt"

func BuildDSN(config Config) string {
	return "postgres://" + config.Username + ":" + config.Password +
		"@" + config.Host + ":" + fmt.Sprint(config.Port) +
		"/" + config.DBName + "?sslmode=" + config.SSLMode
}
