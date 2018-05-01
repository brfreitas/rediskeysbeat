// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period     time.Duration `config:"period"`
	Host       string        `config:"host"`
	Port       int           `config:"port"`
	Network    string        `config:"network"`
	MaxConn    int           `config:"maxconn"`
	Auth       AuthConfig
	KeyPattern string `config:"keypattern"`
	KeyEntity  string `config:"keyentity"`
}

type AuthConfig struct {
	Required     bool   `config:"required"`
	RequiredPass string `config:"requiredpass"`
}

var DefaultConfig = Config{
	Period:  10 * time.Second,
	Host:    "localhost",
	Port:    6379,
	Network: "tcp",
	MaxConn: 10,
	Auth: AuthConfig{
		Required:     false,
		RequiredPass: "",
	},
	KeyPattern: "",
	KeyEntity:  "",
}
