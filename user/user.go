/*
 * Copyright 2014 The starfruit Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package user

import (
	"bufio"
	"fmt"
	"github.com/flatpeach/starfruit/config"
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/version"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	StatusPasswordNotVerified = iota
	StatusPasswordVerified
	StatusNotRegistered
	StatusRegistered
	StatusDisconnecting
)

type Mode int

const (
	ModeAway                     Mode = 1 << 0
	ModeInvisible                Mode = 1 << 1
	ModeReceiveWallops           Mode = 1 << 2
	ModeRestrictedUserConnection Mode = 1 << 3
	ModeOperator                 Mode = 1 << 4
	ModeLocalOperator            Mode = 1 << 5
	ModeReceiveServiceNotice     Mode = 1 << 6
)

func (m Mode) String() string {
	switch m {
	case ModeAway:
		return "away"

	case ModeInvisible:
		return "invisible"

	case ModeReceiveWallops:
		return "receive wallops"

	case ModeRestrictedUserConnection:
		return "restricted user connection"

	case ModeOperator:
		return "operator"

	case ModeLocalOperator:
		return "local operator"

	case ModeReceiveServiceNotice:
		return "receive service notice"

	default:
		return "Unknown"
	}
}

type User struct {
	Config *config.Config // Global Server Config

	Conn net.Conn // Original TCP connection

	Id           int
	UserName     string
	NickName     string
	RealName     string
	HostName     string // Hostname this user try to connect
	LastPongTime int64  // Last time this user reply a PONG message

	In  chan []byte
	Out chan []byte

	awayMsg string // Away message for this user
	status  int    // @Todo: Replace this with real FSM
	modes   Mode
	mutex   sync.Mutex
}

func New(cf *config.Config, conn net.Conn) *User {
	u := &User{
		Config:       cf,
		Conn:         conn,
		status:       StatusPasswordNotVerified,
		LastPongTime: time.Now().Unix(),
		Id:           0,
	}

	if cf.Password == "" {
		u.EnterStatus(StatusPasswordVerified)
	} else {
		u.EnterStatus(StatusPasswordNotVerified)
	}

	u.In = make(chan []byte)
	u.Out = make(chan []byte)

	return u
}

func (u *User) AwayMsg() string {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	return u.awayMsg
}

func (u *User) SetAwayMsg(s string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.awayMsg = s
}

func (u *User) MarkAway(b bool) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if b {
		u.modes |= ModeAway
	} else {
		u.modes = u.modes & ^ModeAway
	}
}

func (u *User) IsAway() bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.modes&ModeAway > 0 {
		return true
	}

	return false
}

func (u *User) MarkMode(m Mode) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.modes |= m
}

func (u *User) ClearMode(m Mode) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.modes &= ^m
}

func (u *User) Modes() string {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	s := "+"

	if u.modes&ModeAway > 0 {
		s += "a"
	}

	if u.modes&ModeInvisible > 0 {
		s += "i"
	}

	if u.modes&ModeReceiveWallops > 0 {
		s += "w"
	}

	if u.modes&ModeRestrictedUserConnection > 0 {
		s += "r"
	}

	if u.modes&ModeOperator > 0 {
		s += "o"
	}

	if u.modes&ModeLocalOperator > 0 {
		s += "O"
	}

	if u.modes&ModeReceiveServiceNotice > 0 {
		s += "s"
	}

	return s
}

func (u *User) IsRegistered() bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.status == StatusRegistered {
		return true
	}

	return false
}

func (u *User) IsDisconnecting() bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.status == StatusDisconnecting {
		return true
	}

	return false
}

func (u *User) Full() string {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	return u.NickName + "!~" + u.UserName + "@" + u.HostName
}

func (u *User) Close() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.status == StatusDisconnecting {
		return
	}

	u.status = StatusDisconnecting

	if u.Conn == nil {
		log.Printf("[SERVER] Try to close nil connection")
		return
	}

	err := u.Conn.Close()
	if err != nil {
		log.Printf("[SERVER] Failed to close user's connection: %s", err)
	}

	close(u.In)
	close(u.Out)
}

func (u *User) IsPasswordVerified() bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.status >= StatusPasswordVerified {
		return true
	}

	return false
}

func (u *User) SendMessage(m *message.Message) {
	if u.IsDisconnecting() {
		return
	}
	if m != nil {
		data := m.String() + "\r\n"
		u.Out <- []byte(data)
	} else {
		u.Out <- nil
	}
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
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.status = s
}

func (u *User) Status() int {
	u.mutex.Lock()
	defer u.mutex.Unlock()

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
