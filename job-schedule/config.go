package main

import (
	"crypto/md5"
	"encoding/hex"
	"log"

	"github.com/olebedev/config"
)

type Job struct {
	Description string
	Identity    string
	Cmd         string
	Params      string
	Max         int
}

type Config struct {
	File  string
	Times int
	Jobs  map[string]Job
}

func GetConfig(path string) *Config {
	c, err := config.ParseYamlFile(path)
	if err != nil {
		log.Println(err.Error())
	}

	times, err := c.Int("times")
	if err != nil {
		log.Println(err.Error())
	}

	l, err := c.List("jobs")
	if err != nil {
		log.Println(err.Error())
	}

	jobs := make(map[string]Job)
	for _, item := range l {
		i, _ := item.(map[string]interface{})
		job := Job{}
		job.Description = i["description"].(string)
		job.Cmd = i["cmd"].(string)
		job.Params = i["params"].(string)

		md5Ctx := md5.New()
		md5Ctx.Write([]byte(job.Params))
		cipherStr := md5Ctx.Sum(nil)
		job.Identity = hex.EncodeToString(cipherStr)

		job.Max = i["max"].(int)
		jobs[job.Identity] = job
	}

	config := &Config{}
	config.File = path
	config.Times = times
	config.Jobs = jobs

	return config
}
