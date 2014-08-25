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
	"strings"
)

type Part struct{}

func (module *Part) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// PART <channel> *( "," <channel> ) [ <Part Message> ]

	if len(m.Params) < 1 {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NEEDMOREPARAMS,
			nil,
			"Need more parameters.",
		))
		return nil
	}

	var partMessage string

	if len(m.Params) > 1 {
		partMessage = m.Params[1]
	}

	channelNames := strings.Split(m.Params[0], ",")

	for _, channelName := range channelNames {
		cnl := s.FindChannelByName(channelName)
		if cnl == nil {
			continue
		}

		if cnl == nil {
			u.SendMessage(message.New(
				s.Config.ServerName,
				message.ERR_NOSUCHCHANNEL,
				[]string{channelName},
				"You are not on that channel.",
			))

			continue
		}

		if !s.IsUserJoinedChannel(u.Id, cnl.Id) {
			u.SendMessage(message.New(
				s.Config.ServerName,
				message.ERR_NOTONCHANNEL,
				[]string{
					u.NickName,
					channelName,
				},
				"You are not on that channel.",
			))

			continue
		}

		cnl.Broadcast(message.New(
			u.Full(),
			"PART",
			[]string{channelName},
			partMessage,
		), nil)

		cnl.Quit(u.Id)

	}

	return nil
}
