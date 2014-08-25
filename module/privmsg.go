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

type Privmsg struct{}

func (module *Privmsg) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// PRIVMSG <target> <text>

	if len(m.Params) == 0 {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NORECIPIENT,
			[]string{
				u.NickName,
			},
			"No recipient given",
		))

		return nil
	}

	if len(m.Params) == 1 && m.Trailing == "" {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NOTEXTTOSEND,
			[]string{u.NickName},
			"No text to send",
		))
	}

	var (
		msgText string
	)

	msgText = m.Params[1]

	targetUser := s.GetUserByNickName(m.Params[0])
	if targetUser != nil {
		// Send msg to specific user
		targetUser.SendMessage(message.New(
			u.Full(),
			"PRIVMSG",
			[]string{targetUser.NickName},
			msgText,
		))

		return nil
	}

	cnl := s.FindChannelByName(m.Params[0])
	if cnl != nil {
		// Send msg to specific channel
		msg := message.New(
			u.Full(),
			"PRIVMSG",
			[]string{
				cnl.String(),
			},
			msgText,
		)
		s.BroadcastMessage(cnl.Id, msg, []int{u.Id}) // @Todo: implement the exclude ids logic
		return nil
	}

	u.SendMessage(message.New(
		s.Config.ServerName,
		message.ERR_NOSUCHNICK,
		[]string{
			u.NickName,
			m.Params[0],
		},
		nil,
	))

	return nil
}
