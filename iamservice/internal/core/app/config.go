package app

import "github.com/mobiletoly/gokatana/katapp"

type Config struct {
	Deployment  string
	Database    katapp.DatabaseConfig
	Credentials CredentialsConfig
	Server      katapp.ServerConfig
	Cache       katapp.CacheConfig
	GCloud      GCloudConfig
}

type CredentialsConfig struct {
	JwtSecret string
}

type GCloudConfig struct {
	Mock        bool
	ServiceJson string
	Email       struct {
		User string
		From string
	}
}
