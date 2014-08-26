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

type Nick struct{}

func (module *Nick) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// NICK <nickname>

	if !u.IsPasswordVerified() {
		return nil
	}

	if len(m.Params) != 1 {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NEEDMOREPARAMS,
			nil,
			"Need more params",
		))

		return nil
	}

	nickName := m.Params[0]

	if u.IsRegistered() {
		if s.IsNickNameRegistered(nickName) {
			u.SendMessage(message.New(
				s.Config.ServerName,
				message.ERR_NICKNAMEINUSE,
				[]string{
					"*",
					u.NickName,
				},
				"Nickname is already in use",
			))
		}

		nickChangedMsg := message.New(
			u.Full(),
			"NICK",
			nil,
			nickName,
		)

		oldNickName := u.NickName
		u.NickName = nickName

		s.RegisterNickName(u.NickName, u)
		s.UnregisterNickName(oldNickName)

		u.SendMessage(nickChangedMsg)

		for _, c := range s.GetJoinedChannels(u.Id) {
			s.BroadcastMessage(c.Id, nickChangedMsg, []int{u.Id})
		}

		return nil
	}

	u.NickName = nickName

	if u.UserName != "" {
		if s.IsNickNameRegistered(u.NickName) {
			u.SendMessage(message.New(
				s.Config.ServerName,
				message.ERR_NICKNAMEINUSE,
				[]string{
					"*",
					u.NickName,
				},
				"Nickname is already in use",
			))

			return nil
		} else {
			u.Id = s.NewUserId()
			s.RegisterUser(u)
			u.EnterStatus(user.StatusRegistered)
			u.SendWelcomeMessage()
		}
	}

	return nil
}
