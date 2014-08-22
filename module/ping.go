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

type Ping struct{}

func (module *Ping) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// PING [SERVER]
	if len(m.Params) != 1 {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NOORIGIN,
			[]string{u.NickName},
			"No origin specified",
		))

		return nil
	}

	server := m.Params[0]

	u.SendMessage(message.New(
		s.Config.ServerName,
		"PONG",
		[]string{s.Config.ServerName},
		server,
	))

	return nil
}
