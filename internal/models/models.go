package models

type (
	Config struct {
		Env                  string `yaml:"env" json:"env" env:"ENV" env-default:"local"`
		Address              string `yaml:"address" json:"address" env:"ADDRESS"`
		DSN                  string `yaml:"dsn" json:"dsn" env:"DSN"`
		AccrualSystemAddress string `yaml:"accrual_system_address" json:"accrual_system_address" env:"ACCRUAL_SYSTEM_ADDRESS"`
	}
)
