package config

type ConfigReader interface {
	GetConfig() (*Config, error)
}
