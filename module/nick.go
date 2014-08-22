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
)

type Nick struct{}

func (module *Nick) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// NICK <nickname>
	if len(m.Params) != 1 {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NEEDMOREPARAMS,
			nil,
			"Need more params",
		))

		return nil
	}

	u.NickName = m.Params[0]

	if u.UserName != "" {
		u.Id = s.NewUserId()
		s.RegisterUser(u)
		u.EnterStatus(user.Registered)
		u.SendWelcomeMessage()
	}

	return nil
}
