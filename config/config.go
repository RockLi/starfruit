/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package config

import (
	"time"
)

type Config struct {
	ServerName string
	Password   string
	SSL        bool

	BindIP   string
	BindPort int

	PingUserInterval int
}

func New() *Config {
	cf := &Config{
		ServerName:       "chat.starfruit.io",
		SSL:              false,
		PingUserInterval: time.Minute * 3,
	}

	return cf
}
