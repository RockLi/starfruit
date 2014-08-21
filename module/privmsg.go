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

type Privmsg struct{}

func (module *Privmsg) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// PRIVMSG <target> <text>

	if len(m.Params) == 0 {
		u.SendMessage(&message.Message{
			Prefix:  s.Config.ServerName,
			Command: message.ERR_NORECIPIENT,
			Params: []string{
				u.NickName,
			},
			Trailing: "No recipient given",
		})

		return nil
	}

	if len(m.Params) == 1 && m.Trailing == "" {
		u.SendMessage(&message.Message{
			Prefix:  s.Config.ServerName,
			Command: message.ERR_NOTEXTTOSEND,
			Params: []string{
				u.NickName,
			},
			Trailing: "No text to send",
		})
	}

	var (
		msgText string
	)

	msgText = m.Params[1]

	targetUser := s.GetUserByNickName(m.Params[0])
	if targetUser != nil {
		// Send msg to specific user
		targetUser.SendMessage(&message.Message{
			Prefix:  u.Full(),
			Command: "PRIVMSG",
			Params: []string{
				targetUser.NickName,
			},
			Trailing: msgText,
		})
		return nil
	}

	cnl := s.FindChannelByName(m.Params[0])
	if cnl != nil {
		// Send msg to specific channel
		msg := &message.Message{
			Prefix:  u.Full(),
			Command: "PRIVMSG",
			Params: []string{
				cnl.String(),
			},
			Trailing: msgText,
		}
		s.BroadcastMessage(cnl.Id, msg, []int{u.Id}) // @Todo: implement the exclude ids logic
		return nil
	}

	u.SendMessage(&message.Message{
		Prefix:  s.Config.ServerName,
		Command: message.ERR_NOSUCHNICK,
		Params: []string{
			u.NickName,
			m.Params[0],
		},
	})

	return nil
}
