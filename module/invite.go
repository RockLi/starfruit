/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/server"
	"github.com/flatpeach/starfruit/user"
)

type Invite struct{}

func (module *Invite) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// INVITE <nickname> <channel>

	if len(m.Params) != 2 {
		u.SendErrorNeedMoreParams("INVITE")
		return nil
	}

	nick := m.Params[0]
	channelName := m.Params[1]

	invitedUser := s.GetUserByNickName(nick)
	if invitedUser == nil {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NOSUCHNICK,
			[]string{
				u.NickName,
				nick,
			},
			"No such nick",
		))

		return nil
	}

	c := s.FindChannelByName(channelName)
	if c != nil {
		if !s.IsUserJoinedChannel(u.Id, c.Id) {
			u.SendMessage(message.New(
				s.Config.ServerName,
				message.ERR_NOTONCHANNEL,
				[]string{
					u.NickName,
					channelName,
				},
				"You're not on that channel",
			))

			return nil
		} else if s.IsUserJoinedChannel(invitedUser.Id, c.Id) {
			u.SendMessage(message.New(
				s.Config.ServerName,
				message.ERR_USERONCHANNEL,
				[]string{
					u.NickName,
					nick,
					channelName,
				},
				"is already on channel",
			))

			return nil
		}
	}

	u.SendMessage(message.New(
		s.Config.ServerName,
		message.RPL_INVITING,
		[]string{
			u.NickName,
			invitedUser.NickName,
			channelName,
		},
		nil,
	))

	invitedUser.SendMessage(message.New(
		u.Full(),
		"INVITE",
		[]string{
			invitedUser.NickName,
		},
		channelName,
	))

	if invitedUser.IsAway() {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.RPL_AWAY,
			[]string{
				u.NickName,
				invitedUser.NickName,
			},
			invitedUser.AwayMsg(),
		))
	}

	return nil
}
