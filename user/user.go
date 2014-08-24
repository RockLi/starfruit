/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package user

import (
	"bufio"
	"fmt"
	"github.com/flatpeach/ircd/config"
	"github.com/flatpeach/ircd/message"
	"github.com/flatpeach/ircd/version"
	"log"
	"net"
	"os"
	"time"
)

const (
	PasswordNotVerified = iota
	PasswordVerified
	NotRegistered
	Registered
	Disconnecting
)

type User struct {
	Config *config.Config // Global Server Config
	Conn   net.Conn       // Original TCP connection
	status int            // @Todo: Replace this with real FSM

	LastPongTime int64 // Last time this user reply a PONG message

	In  chan []byte
	Out chan []byte

	Id       int
	UserName string
	NickName string
	RealName string
	Mode     int32
	HostName string // Hostname this user try to connect
	Away     string // Away message for this user
}

func New(cf *config.Config, conn net.Conn) *User {
	u := &User{
		Config:       cf,
		Conn:         conn,
		status:       PasswordNotVerified,
		LastPongTime: time.Now().Unix(),
		Id:           0,
	}

	if cf.Password == "" {
		u.EnterStatus(PasswordVerified)
	} else {
		u.EnterStatus(PasswordNotVerified)
	}

	u.In = make(chan []byte)
	u.Out = make(chan []byte)

	return u
}

func (u *User) IsRegistered() bool {
	if u.status == Registered {
		return true
	}

	return false
}

func (u *User) IsDisconnecting() bool {
	if u.status == Disconnecting {
		return true
	}

	return false
}

func (u *User) Full() string {
	return u.NickName + "!~" + u.UserName + "@" + u.HostName
}

func (u *User) Close() {
	if u.status == Disconnecting {
		return
	}
	if u.Conn == nil {
		log.Printf("[SERVER] Try to close nil connection")
		return
	}

	err := u.Conn.Close()
	if err != nil {
		log.Printf("[SERVER] Failed to close user's connection: %s", err)
	}

	// close(u.In)
	// close(u.Out)
}

func (u *User) IsPasswordVerified() bool {
	if u.status >= PasswordVerified {
		return true
	}

	return false
}

func (u *User) SendMessage(m *message.Message) {
	data := m.String() + "\r\n"
	u.Out <- []byte(data)
}

func (u *User) SendErrorNeedMoreParams(c string) {
	m := message.New(
		u.Config.ServerName,
		message.ERR_NEEDMOREPARAMS,
		[]string{
			u.NickName,
			c,
		},
		"Not enough parameters",
	)

	if u.NickName == "" {
		m.Params[0] = "*"
	}

	u.SendMessage(m)
}

func (u *User) EnterStatus(s int) {
	u.status = s
}

func (u *User) Status() int {
	return u.status
}

func (u *User) SendMotd() {

	file, err := os.Open(u.Config.MotdFile)
	if err != nil {
		u.SendMessage(message.New(
			u.Config.ServerName,
			message.ERR_NOMOTD,
			nil,
			"MOTD File is missing",
		))

		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	u.SendMessage(message.New(
		u.Config.ServerName,
		message.RPL_MOTDSTART,
		[]string{u.NickName},
		fmt.Sprintf("- %s Message of the day -", u.Config.ServerName),
	))

	for {
		buf, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		u.SendMessage(message.New(
			u.Config.ServerName,
			message.RPL_MOTD,
			[]string{u.NickName},
			fmt.Sprintf("- %s -", buf),
		))
	}

	u.SendMessage(message.New(
		u.Config.ServerName,
		message.RPL_ENDOFMOTD,
		[]string{u.NickName},
		"End of /MOTD command.",
	))
}

func (u *User) SendWelcomeMessage() {
	u.SendMessage(message.New(
		u.Config.ServerName,
		message.RPL_WELCOME,
		[]string{u.NickName},
		fmt.Sprintf("Welcome to %s", u.Config.ServerName),
	))

	u.SendMessage(message.New(
		u.Config.ServerName,
		message.RPL_YOURHOST,
		[]string{u.NickName},
		fmt.Sprintf("Your host is %s, running version %s",
			u.Config.ServerName,
			version.Version(),
		),
	))

	u.SendMessage(message.New(
		u.Config.ServerName,
		message.RPL_CREATED,
		[]string{u.NickName},
		fmt.Sprintf("This server was created %s",
			u.Config.ServerCreatedAt.Format("Jan 2, 2006 at 3:04pm (MST)")),
	))

	u.SendMessage(message.New(
		u.Config.ServerName,
		message.RPL_MYINFO,
		[]string{u.NickName},
		fmt.Sprintf("%s %s", u.Config.ServerName, version.Version()),
	))

	// u.SendMessage(message.New(
	// 	u.Config.ServerName,
	// 	message.RPL_BOUNCE,
	// 	[]string{u.NickName},
	// 	nil,
	// ))

	u.SendMotd()

	u.SendMessage(message.New(
		u.Full(),
		"MODE",
		[]string{u.NickName},
		"+i",
	))
}
