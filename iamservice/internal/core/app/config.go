package app

import "github.com/mobiletoly/gokatana/katapp"

type Config struct {
	Deployment  string
	Database    katapp.DatabaseConfig
	Credentials CredentialsConfig
	Server      katapp.ServerConfig
	Cache       katapp.CacheConfig
}

type CredentialsConfig struct {
	Secret string
}
