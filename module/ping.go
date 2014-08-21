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
	if len(m.Params) == 0 && m.Trailing == "" {
		u.SendMessage(&message.Message{
			Prefix:   s.Config.ServerName,
			Command:  message.ERR_NOORIGIN,
			Params:   []string{u.NickName},
			Trailing: "No origin specified",
		})
		return nil
	}

	var server string

	if m.Trailing != "" {
		server = m.Trailing
	} else {
		server = m.Params[0]
	}

	u.SendMessage(&message.Message{
		Prefix:   s.Config.ServerName,
		Command:  "PONG",
		Params:   []string{s.Config.ServerName},
		Trailing: server,
	})

	return nil
}
