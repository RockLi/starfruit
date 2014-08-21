/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package user

import (
	"fmt"
	"github.com/flatpeach/ircd/config"
	"github.com/flatpeach/ircd/message"
	"log"
	"net"
)

type User struct {
	Config *config.Config // Global Server Config
	Conn   net.Conn       // Original TCP connection
	Status string         // @Todo: Replace this one with FSM

	LastPongTime int64 // Last time this user reply a PONG message

	In chan []byte

	Id int

	UserName string
	NickName string
	RealName string
	Mode     int32
	HostName string // Hostname this user try to connect
	Away     string // Away message for this user
}

func New(cf *config.Config, conn net.Conn) *User {
	u := &User{
		Config: cf,
		Conn:   conn,
		// stateMachine: fsm.New(), // Manipulating User State
		LastPongTime: 0,
	}

	if cf.Password == "" {
		u.Status = "PasswordVerified"
	} else {
		u.Status = "WaitingPassword"
	}

	u.In = make(chan []byte)

	return u
}

func (u *User) Full() string {
	return u.NickName + "!~" + u.UserName + "@" + u.HostName
}

func (u *User) Close() {
	if u.Conn == nil {
		log.Printf("[SERVER] Try to close nil connection")
		return
	}

	err := u.Conn.Close()
	if err != nil {
		log.Printf("[SERVER] Failed to close user's connection: %s", err)
	}
}

func (u *User) PasswordVerified() bool {
	if u.Status != "WaitingPassword" {
		return true
	}

	return false
}

func (u *User) SendMessage(m *message.Message) {
	data := m.String() + "\r\n"
	log.Printf("[Client:%s] Reply %s", u.Conn.RemoteAddr(), m.String())
	_, err := u.Conn.Write([]byte(data))
	if err != nil {
		log.Printf("[Client:%s] Failed to send reply message")
		u.Close()
	}
}

func (u *User) SendErrorNeedMoreParams(c string) {
	m := &message.Message{
		Prefix:  u.Config.ServerName,
		Command: message.ERR_NEEDMOREPARAMS,
		Params: []string{
			u.NickName,
			c,
		},
		Trailing: "Not enough parameters",
	}

	if u.NickName == "" {
		m.Params[0] = "*"
	}

	u.SendMessage(m)
}

func (u *User) SendWelcomeMessage() {
	u.SendMessage(&message.Message{
		Prefix:   u.Config.ServerName,
		Command:  message.RPL_WELCOME,
		Params:   []string{u.NickName},
		Trailing: fmt.Sprintf("Welcome to %s", u.Config.ServerName),
	})

	u.SendMessage(&message.Message{
		Prefix:   u.Config.ServerName,
		Command:  "002",
		Params:   []string{u.NickName},
		Trailing: "xxx",
	})

	u.SendMessage(&message.Message{
		Prefix:   u.Config.ServerName,
		Command:  "003",
		Params:   []string{u.NickName},
		Trailing: "xxx",
	})

	u.SendMessage(&message.Message{
		Prefix:   u.Config.ServerName,
		Command:  "004",
		Params:   []string{u.NickName},
		Trailing: "xxx",
	})

	u.SendMessage(&message.Message{
		Prefix:   u.Config.ServerName,
		Command:  "005",
		Params:   []string{u.NickName},
		Trailing: "xxx",
	})

	u.SendMessage(&message.Message{
		Prefix:   u.Config.ServerName,
		Command:  "375",
		Params:   []string{u.NickName},
		Trailing: "im.starfruit.io Message of the day",
	})

	u.SendMessage(&message.Message{
		Prefix:   u.Config.ServerName,
		Command:  "372",
		Params:   []string{u.NickName},
		Trailing: "good",
	})

	u.SendMessage(&message.Message{
		Prefix:   u.Config.ServerName,
		Command:  message.RPL_ENDOFMOTD,
		Params:   []string{u.NickName},
		Trailing: "End of /MOTD command.",
	})

	u.SendMessage(&message.Message{
		Prefix:   u.Full(),
		Command:  "MODE",
		Params:   []string{u.NickName},
		Trailing: "+i",
	})
}
