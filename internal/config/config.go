package config

import (
	"fmt"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/ilyakaznacheev/cleanenv"
)

func MustConfig() models.Config {
	cfg := getConfigByEnv()

	return cfg
}

func getConfigByEnv() models.Config {
	const fn = "config.getConfigByEnv"
	var cfg models.Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Sprintf("%s: %v", fn, err))
	}

	return cfg
}
