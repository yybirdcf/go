package main

import (
	"log"

	"github.com/olebedev/config"
)

type Config struct {
	DB struct {
		Host string
	}

	Redis struct {
		Host string
		Pwd  string
		Db   int
	}

	Https struct {
		Addr   string
		Enable bool
		Cert   string
		Key    string
	}

	Debug string
	Admin string
	Proxy string
	Etcd  []string

	Log struct {
		Access string
		Error  string
	}

	Limit struct {
		TTL int
		Max int
	}
}

func GetConfig(path string) *Config {
	c, err := config.ParseYamlFile(path)
	if err != nil {
		log.Fatalf("error parse yaml config file: %s\n", err.Error())
	}
	db_host, err := c.String("db.host")
	if err != nil {
		log.Fatalf("error db.host: %s\n", err.Error())
	}

	redis_host, err := c.String("redis.host")
	if err != nil {
		log.Fatalf("error redis.host: %s\n", err.Error())
	}

	redis_pwd, err := c.String("redis.pwd")
	if err != nil {
		log.Fatalf("error redis.pwd: %s\n", err.Error())
	}

	redis_db, err := c.Int("redis.db")
	if err != nil {
		log.Fatalf("error redis.db: %s\n", err.Error())
	}

	debug, err := c.String("debug")
	if err != nil {
		log.Fatalf("error debug: %s\n", err.Error())
	}

	admin, err := c.String("admin")
	if err != nil {
		log.Fatalf("error admin: %s\n", err.Error())
	}

	proxy, err := c.String("proxy")
	if err != nil {
		log.Fatalf("error proxy: %s\n", err.Error())
	}

	l, err := c.List("etcd")
	if err != nil {
		log.Fatalf("error etcd: %s\n", err.Error())
	}

	etcdl := []string{}
	for _, item := range l {
		if t, ok := item.(string); ok {
			etcdl = append(etcdl, t)
		}
	}

	https_enable, err := c.Bool("https.enable")
	if err != nil {
		log.Fatalf("error https.enable: %s\n", err.Error())
	}
	https_addr := ""
	https_cert := ""
	https_key := ""
	if https_enable {
		https_addr, err = c.String("https.addr")
		if err != nil {
			log.Fatalf("error https.addr: %s\n", err.Error())
		}
		https_cert, err = c.String("https.addr")
		if err != nil {
			log.Fatalf("error https.cert: %s\n", err.Error())
		}
		https_key, err = c.String("https.key")
		if err != nil {
			log.Fatalf("error https.key: %s\n", err.Error())
		}
	}

	log_access, err := c.String("log.access")
	if err != nil {
		log.Fatalf("error log.access: %s\n", err.Error())
	}
	log_error, err := c.String("log.error")
	if err != nil {
		log.Fatalf("error log.error: %s\n", err.Error())
	}

	limit_ttl, err := c.Int("limit.ttl")
	if err != nil {
		log.Fatalf("error limit.ttl: %s\n", err.Error())
	}
	limit_max, err := c.Int("limit.max")
	if err != nil {
		log.Fatalf("error limit.max: %s\n", err.Error())
	}

	cfg := &Config{}
	cfg.DB.Host = db_host
	cfg.Debug = debug
	cfg.Admin = admin
	cfg.Redis.Db = redis_db
	cfg.Redis.Host = redis_host
	cfg.Redis.Pwd = redis_pwd
	cfg.Proxy = proxy
	cfg.Etcd = etcdl
	cfg.Https.Addr = https_addr
	cfg.Https.Enable = https_enable
	cfg.Https.Cert = https_cert
	cfg.Https.Key = https_key
	cfg.Limit.TTL = limit_ttl
	cfg.Limit.Max = limit_max
	cfg.Log.Access = log_access
	cfg.Log.Error = log_error

	return cfg
}
