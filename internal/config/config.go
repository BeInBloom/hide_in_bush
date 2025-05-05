package config

import (
	"flag"
	"fmt"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/ilyakaznacheev/cleanenv"
)

func MustConfig() models.Config {
	cfg := getConfigByEnv()
	parseFlags(&cfg)
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

func parseFlags(cfg *models.Config) {
	runAddressFlag := flag.String("a", "", "Address to run the server")
	databaseDNSFlag := flag.String("d", "", "Address to database")
	accrualSystemAddressFlag := flag.String("r", "", "Address to accrual system")

	flag.Parse()

	if *runAddressFlag != "" {
		cfg.Server.Address = *runAddressFlag
	}

	if *databaseDNSFlag != "" {
		cfg.Server.DSN = *databaseDNSFlag
	}

	if *accrualSystemAddressFlag != "" {
		cfg.Server.AccrualSystemAddress = *accrualSystemAddressFlag
	}
}
