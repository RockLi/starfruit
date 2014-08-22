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

type Away struct{}

func (module *Away) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// AWAY [param]
	var (
		awayMsg string
	)

	if len(m.Params) == 1 {
		awayMsg = m.Params[0]
	}

	u.Away = awayMsg

	if u.Away == "" {
		u.SendMessage(message.New(s.Config.ServerName,
			message.RPL_UNAWAY,
			[]string{u.NickName},
			"You are no longer marked as being away",
		))
	} else {
		u.SendMessage(message.New(s.Config.ServerName,
			message.RPL_NOWAWAY,
			[]string{u.NickName},
			"You have been marked as being away",
		))
	}

	return nil
}
