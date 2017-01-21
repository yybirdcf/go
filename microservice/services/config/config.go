package config

type Config struct {
	DB_app            string
	REDIS_HOST_app    string
	REDIS_PWD_app     string
	REDIS_DB_app      int
	REDIS_HOST_pubsub string
	REDIS_PWD_pubsub  string
	REDIS_DB_pubsub   int
}

func GetConfig(env string) *Config {
	if env == "dev" {
		return &Config{
			DB_app:            "", //user:password@tcp(host:port)/db?charset=utf8
			REDIS_HOST_pubsub: "",
			REDIS_PWD_pubsub:  "",
			REDIS_DB_pubsub:   1,
			REDIS_HOST_app:    "127.0.0.1:6379",
			REDIS_PWD_app:     "",
			REDIS_DB_app:      1,
		}
	} else if env == "test" {
		return &Config{
			DB_app:            "",
			REDIS_HOST_pubsub: "",
			REDIS_PWD_pubsub:  "",
			REDIS_DB_pubsub:   1,
			REDIS_HOST_app:    "127.0.0.1:6379",
			REDIS_PWD_app:     "",
			REDIS_DB_app:      1,
		}
	} else if env == "pro" {
		return &Config{
			DB_app:            "",
			REDIS_HOST_pubsub: "",
			REDIS_PWD_pubsub:  "",
			REDIS_DB_pubsub:   1,
			REDIS_HOST_app:    "127.0.0.1:6379",
			REDIS_PWD_app:     "",
			REDIS_DB_app:      1,
		}
	}

	return &Config{}
}
