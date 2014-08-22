/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"github.com/flatpeach/ircd/message"
	"github.com/flatpeach/ircd/server"
	"github.com/flatpeach/ircd/user"
	"strconv"
)

type User struct{}

func (module *User) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// USER <user> <mode> <unused> <realname>

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
		if m.Params[1] == "*" {
			u.Mode = 0
		} else {
			mode, _ := strconv.Atoi(m.Params[1])
			u.Mode = int32(mode)
		}
		u.HostName = m.Params[2] // @Todo: fix the hostname here
		u.RealName = m.Params[3]
	}

	if u.NickName != "" {
		// Everything is ok, register this user to the server user list
		u.Id = s.NewUserId()
		s.RegisterUser(u)
		u.EnterStatus(user.Registered)
		u.SendWelcomeMessage()
	}

	return nil
}
