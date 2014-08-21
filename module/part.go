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
	"strings"
)

type Part struct{}

func (module *Part) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// PART <channel> *( "," <channel> ) [ <Part Message> ]

	if len(m.Params) < 1 {
		u.SendMessage(&message.Message{
			Prefix:   s.Config.ServerName,
			Command:  message.ERR_NEEDMOREPARAMS,
			Params:   nil,
			Trailing: "Need more parameters.",
		})
		return nil
	}

	var partMessage string

	if m.Trailing != "" {
		partMessage = m.Trailing
	} else if len(m.Params) > 1 {
		partMessage = m.Params[1]
	}

	channelNames := strings.Split(m.Params[0], ",")

	for _, channelName := range channelNames {
		cnl := s.FindChannelByName(channelName)
		if cnl == nil {
			continue
		}

		if cnl == nil {
			u.SendMessage(&message.Message{
				Prefix:  s.Config.ServerName,
				Command: message.ERR_NOSUCHCHANNEL,
				Params: []string{
					channelName,
				},
				Trailing: "You are not on that channel.",
			})
			continue
		}

		if !s.IsUserJoinedChannel(u.Id, cnl.Id) {
			u.SendMessage(&message.Message{
				Prefix:  s.Config.ServerName,
				Command: message.ERR_NOTONCHANNEL,
				Params: []string{
					u.NickName,
					channelName,
				},
				Trailing: "You are not on that channel.",
			})
			continue
		}

		cnl.Broadcast(&message.Message{
			Prefix:  u.Full(),
			Command: "PART",
			Params: []string{
				channelName,
			},
			Trailing: partMessage,
		}, nil)

		cnl.Quit(u.Id)

	}

	return nil
}
