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

	if !(len(m.Params) == 4 || (len(m.Params) == 3 && m.Trailing != "")) {
		u.SendMessage(&message.Message{
			Prefix:   s.Config.ServerName,
			Command:  message.ERR_NEEDMOREPARAMS,
			Params:   nil,
			Trailing: "Need more params",
		})

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
		u.HostName = m.Params[2]
		if u.HostName == "" {
			// @Todo: fix hostname
		}

		if m.Trailing != "" {
			u.RealName = m.Trailing
		} else {
			u.RealName = m.Params[3]
		}
	}

	if u.NickName != "" {
		// Everything is ok, register this user to the server user list
		u.Id = s.NewUserId()
		s.RegisterUser(u)

		u.SendWelcomeMessage()
	}

	return nil
}
