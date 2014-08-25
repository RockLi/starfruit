/*
 * Copyright 2014 The starfruit Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package config

import (
	"encoding/json"
	"io/ioutil"
	_ "strconv"
	"time"
)

type Config struct {
	ServerName      string
	ServerCreatedAt time.Time
	Password        string
	SSL             bool

	BindIP   string
	BindPort int

	PingUserInterval int
	UserTimeout      int

	MotdFile string
}

func New() *Config {
	cf := &Config{
		ServerName:       "localhost",
		ServerCreatedAt:  time.Date(2014, 01, 01, 12, 0, 0, 0, time.UTC),
		SSL:              false,
		BindIP:           "127.0.0.1",
		BindPort:         6667,
		PingUserInterval: 120,
		UserTimeout:      300,
		MotdFile:         "",
	}

	return cf
}

func (c *Config) LoadFromJSONFile(name string) error {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	var configs map[string]interface{}

	err = json.Unmarshal(data, &configs)
	if err != nil {
		return err
	}

	if ip, exists := configs["bind_ip"]; exists {
		c.BindIP = ip.(string)
	}

	if port, exists := configs["bind_port"]; exists {
		c.BindPort = int(port.(float64))
	}

	if serverName, exists := configs["server_name"]; exists {
		c.ServerName = serverName.(string)
	}

	if password, exists := configs["password"]; exists {
		c.Password = password.(string)
	}

	if enableSSL, exists := configs["enable_ssl"]; exists {
		c.SSL = enableSSL.(bool)
	}

	if pingUserInterval, exists := configs["ping_user_interval"]; exists {
		c.PingUserInterval = int(pingUserInterval.(float64))
	}

	if userTimeout, exists := configs["user_timeout"]; exists {
		c.UserTimeout = int(userTimeout.(float64))
	}

	if motd, exists := configs["motd_file"]; exists {
		c.MotdFile = motd.(string)
	}

	return nil
}
