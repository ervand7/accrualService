package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
)

const OrdersBatchSize int = 5

var (
	runAddressFlag           *string
	databaseURIFlag          *string
	accrualSystemAddressFlag *string
)

func init() {
	runAddressFlag = flag.String("a", "", "Main server run address")
	databaseURIFlag = flag.String("d", "", "Database source name")
	accrualSystemAddressFlag = flag.String("r", "", "Accrual system address")
}

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:":8081"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://127.0.0.1:8080"`
}

// GetConfig flag value has more priority than env value
func GetConfig() Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	flag.Parse()
	if *runAddressFlag != "" {
		cfg.RunAddress = *runAddressFlag
	}
	if *databaseURIFlag != "" {
		cfg.DatabaseURI = *databaseURIFlag
	}
	if *accrualSystemAddressFlag != "" {
		cfg.AccrualSystemAddress = *accrualSystemAddressFlag
	}

	return cfg
}
