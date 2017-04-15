package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/howeyc/fsnotify"
)

type Schedule struct {
	cfg      *Config
	handlers map[string]struct{}
	mutex    sync.Mutex
}

func NewSchedule(cfg *Config) *Schedule {
	s := &Schedule{
		cfg:      cfg,
		handlers: make(map[string]struct{}),
	}

	s.monitor()

	return s
}

func (s *Schedule) monitor() {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}

	err = watch.Watch(s.cfg.File)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		for {
			select {
			case w := <-watch.Event:
				if w.IsModify() {
					s.reload()
				}
			case err := <-watch.Error:
				log.Println(err)
			}
		}
	}()
}

func (s *Schedule) reload() {
	defer func() {
		if re := recover(); re != nil {
			stack := make([]byte, 4*1024)
			stack = stack[:runtime.Stack(stack, false)]
			log.Println(string(stack))
		}
	}()

	cfg := GetConfig(s.cfg.File)
	s.cfg = cfg
	s.Run()
}

func (s *Schedule) Run() {
	//添加新增的
	for identity, _ := range s.cfg.Jobs {
		if _, ok := s.handlers[identity]; ok {
			continue
		}

		go s.runJob(identity)

		s.mutex.Lock()
		s.handlers[identity] = struct{}{}
		s.mutex.Unlock()
	}

	//删除老的
	for identity, _ := range s.handlers {
		if _, ok := s.cfg.Jobs[identity]; ok {
			continue
		}

		s.mutex.Lock()
		delete(s.handlers, identity)
		s.mutex.Unlock()
	}
}

func (s *Schedule) runJob(identity string) {
	for {
		time.Sleep(time.Duration(time.Millisecond * 500))

		if _, ok := s.handlers[identity]; !ok {
			return
		}

		if job, ok := s.cfg.Jobs[identity]; ok {
			n := count(identity)
			if n == -1 {
				//放弃，记录
				continue
			}

			if n >= job.Max {
				continue
			}

			params := fmt.Sprintf("%s %s", job.Params, job.Identity)
			args := strings.Split(params, " ")
			cmd := exec.Command(job.Cmd, args...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err := cmd.Start(); err != nil {
				log.Println("Start: ", err.Error())
				time.Sleep(time.Duration(time.Millisecond * 500))
			}
		}
	}
}

func count(c string) int {
	s := fmt.Sprintf("ps aux|grep %s|grep -v grep|wc -l", c)
	cmd := exec.Command("/bin/sh", "-c", s)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("StdoutPipe: ", err.Error())
		return -1
	}

	if err := cmd.Start(); err != nil {
		log.Println("Start: ", err.Error())
		return -1
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Println("ReadAll stdout: ", err.Error())
		return -1
	}

	if err := cmd.Wait(); err != nil {
		log.Println("Wait: ", err.Error())
		return -1
	}

	i := strings.TrimSpace(strings.Trim(string(bytes), "\n"))
	count, err := strconv.Atoi(i)
	if err != nil {
		log.Println("string to int: ", err.Error())
		return -1
	}

	return count
}
