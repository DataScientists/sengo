package testutil

import (
	"sheng-go-backend/config"
	"sheng-go-backend/pkg/util/environment"
)

// REacconfig reads config file for test.

func ReadConfig() {
	config.ReadConfig(config.ReadConfigOption{
		AppEnv: environment.Test,
	})
}

func ReadConfigE2E() {
	config.ReadConfig(config.ReadConfigOption{
		AppEnv: environment.E2E,
	})
}
