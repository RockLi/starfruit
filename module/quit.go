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

type Quit struct{}

func (module *Quit) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// QUIT [ <Quit Message> ]

	var quitMessage string

	if len(m.Params) > 1 {
		quitMessage = m.Params[0]
	}

	u.SendMessage(message.New(
		nil,
		"ERROR",
		nil,
		quitMessage,
	))

	quitMsg := message.New(
		u.Full(),
		"QUIT",
		nil,
		quitMessage,
	)

	channels := s.GetJoinedChannels(u.Id)
	for _, cnl := range channels {
		cnl.Broadcast(quitMsg, nil)
		cnl.Quit(u.Id)
	}

	s.RemoveUser(u.Id)

	return nil
}
