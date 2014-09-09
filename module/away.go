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

	u.SetAwayMsg(awayMsg)

	if u.AwayMsg() == "" {
		u.SendMessage(message.New(s.Config.Server.Name,
			message.RPL_UNAWAY,
			[]string{u.NickName},
			"You are no longer marked as being away",
		))
		u.MarkAway(false)
	} else {
		u.SendMessage(message.New(s.Config.Server.Name,
			message.RPL_NOWAWAY,
			[]string{u.NickName},
			"You have been marked as being away",
		))
		u.MarkAway(true)
	}

	return nil
}
