/*
 * Copyright 2014 The starfruit Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package config

import (
	"code.google.com/p/gcfg"
	"errors"
	"fmt"
	//"io/ioutil"
	"strconv"
	"strings"
	//"time"
)

type Ports []int

func (p *Ports) String() string {
	return fmt.Sprint(*p)
}

func (p *Ports) Set(value string) error {
	if len(*p) > 0 {
		return errors.New("[pineapple] Ports already set")
	}

	for _, v := range strings.Split(value, ",") {
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		*p = append(*p, i)
	}

	return nil
}

type Server struct {
	Ip               string   `gcfg:"ip"`               // IP to bind, normally 0.0.0.0
	Ports            []int    `gcfg:"port"`             // Port to bind
	Name             string   `gcfg:"name"`             // Name of this IRC server
	CreatedAt        string   `gcfg:"created"`          // Creted Time of this IRC server
	SSL              bool     `gcfg:"ssl"`              // Enable SSL or not
	Password         string   `gcfg:"password"`         // Password of this IRC server
	CertFile         string   `gcfg:"cert-file"`        // Cert for SSL
	KeyFile          string   `gcfg:"key-file"`         // Key for SSL
	DisabledCommands []string `gcfg:"disabled-command"` // Which commands to disable
}

type Motd struct {
	File string
}

type Recycle struct {
	PingInterval int `gcfg:"ping-interval"`
	UserTimeout  int `gcfg:"user-timeout"`
}

type Config struct {
	Server  Server
	Motd    Motd
	Recycle Recycle
}

func New() *Config {
	cf := &Config{
		Server: Server{
			Ip:               "127.0.0.1",
			Ports:            []int{6667},
			Name:             "irc.starfruit.io",
			CreatedAt:        "xxx", // @Fix this
			SSL:              false,
			Password:         "",
			CertFile:         "",
			KeyFile:          "",
			DisabledCommands: []string{},
		},
		Motd: Motd{File: ""},
		Recycle: Recycle{
			PingInterval: 300,
			UserTimeout:  300,
		},
	}
	return cf
}

func (c *Config) LoadFromFile(name string) error {
	err := gcfg.ReadFileInto(c, name)
	return err
}
