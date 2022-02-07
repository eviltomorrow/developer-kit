// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Simple service that only works by printing a log message every few seconds.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/zlog"
	"github.com/kardianos/service"
	"go.uber.org/zap"
)

var logger service.Logger

// Program structures.
//  Define Start and Stop methods.
type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	if service.Interactive() {
		logger.Info("Running in terminal.")
	} else {
		logger.Info("Running under service manager.")
	}
	p.exit = make(chan struct{})

	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() error {
	logger.Infof("I'm running %v.", service.Platform())
	path, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("get current dir failure, nest error: %v", err)
	}

	ticker := time.NewTicker(20 * time.Millisecond)
	for {
		select {
		case tm := <-ticker.C:
			logger.Infof("Still running at %v...", tm)
			dosomething(path)
		case <-p.exit:
			ticker.Stop()
			return nil
		}
	}
}

func dosomething(path string) {
	zlog.Info("Hello world, hahahahha", zap.String("path", path))
}

func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.
	logger.Info("I'm Stopping!")
	close(p.exit)
	return nil
}

// Service setup.
//   Define service config.
//   Create the service.
//   Setup the logger.
//   Handle service controls (optional).
//   Run the service.
func main() {
	global, prop, err := zlog.InitLogger(&zlog.Config{
		Level:            "info",
		Format:           "text",
		DisableTimestamp: false,
		File: zlog.FileLogConfig{
			Filename: "C:/data/data.log",
			MaxSize:  1,
		},
	})
	if err != nil {
		fmt.Printf("配置日志信息错误，nest error: %v\r\n", err)
		os.Exit(1)
	}
	zlog.ReplaceGlobals(global, prop)

	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	svcConfig := &service.Config{
		Name:        "GoServiceExampleLogging",
		DisplayName: "GoService",
		Description: "This is an example Go service that outputs log messages.",
		// Dependencies: []string{
		// 	"Requires=network.target",
		// 	"After=network-online.target syslog.target"},
		Option: options,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
