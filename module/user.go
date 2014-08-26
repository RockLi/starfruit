/*
 * Copyright 2014 The starfruit Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/server"
	"github.com/flatpeach/starfruit/user"
	"strconv"
	"strings"
)

type User struct{}

func (module *User) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// USER <user> <mode> <unused> <realname>

	if !u.IsPasswordVerified() {
		return nil
	}

	if !(len(m.Params) == 4) {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NEEDMOREPARAMS,
			nil,
			"Need more params",
		))

		return nil
	}

	if u.UserName == "" {
		u.UserName = m.Params[0]
		mode := m.Params[1]
		if mode != "*" {
			mode, err := strconv.Atoi(mode)
			if err == nil {
				if mode&2 > 0 {
					u.MarkMode(user.ModeReceiveWallops)
				}

				if mode&4 > 0 {
					u.MarkMode(user.ModeInvisible)
				}
			}
		}

		u.HostName = strings.Split(u.Conn.RemoteAddr().String(), ":")[0]
		u.RealName = m.Params[3]
	}

	if u.NickName != "" {
		// Everything is ok, register this user to the server user list
		u.Id = s.NewUserId()
		s.RegisterUser(u)
		u.EnterStatus(user.StatusRegistered)
		u.SendWelcomeMessage()
	}

	return nil
}
