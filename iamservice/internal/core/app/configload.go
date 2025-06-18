package app

import (
	"github.com/mobiletoly/gokatana/katapp"
)

// This prefix is used for environment variables to override config file values
const envVarPrefix = "IAMSERVICE"

func LoadConfig(deployment string) *Config {
	var commonDir string
	if deployment == "test" {
		commonDir = "../configs"
	}
	deploymentConfig := katapp.Deployment{
		Name:            deployment,
		CommonConfigDir: commonDir,
	}
	cfg := katapp.LoadConfig[Config](envVarPrefix, deploymentConfig)
	cfg.Deployment = deployment
	return cfg
}
