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

type Topic struct{}

func (module *Topic) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// TOPIC <channel> [ <topic> ]

	if len(m.Params) == 0 {
		u.SendErrorNeedMoreParams("TOPIC")

		return nil
	}

	channelName := m.Params[0]

	cnl := s.FindChannelByName(channelName)
	if cnl == nil {
		u.SendMessage(&message.Message{
			Prefix:   s.Config.ServerName,
			Command:  message.ERR_NOSUCHCHANNEL,
			Params:   []string{u.NickName, channelName},
			Trailing: "No such channel",
		})
		return nil
	}

	if !s.IsUserJoinedChannel(u.Id, cnl.Id) {
		u.SendMessage(&message.Message{
			Prefix:   s.Config.ServerName,
			Command:  message.ERR_NOTONCHANNEL,
			Params:   []string{u.NickName, channelName},
			Trailing: "You're not on that channel",
		})
		return nil
	}

	if len(m.Params) > 1 || m.Trailing != "" {
		var newTopic string
		if len(m.Params) > 1 {
			newTopic = m.Params[1]
		} else {
			newTopic = m.Trailing
		}

		cnl.SetTopic(newTopic)

		s.BroadcastMessage(cnl.Id, &message.Message{
			Prefix:  u.Full(),
			Command: "TOPIC",
			Params: []string{
				channelName,
			},
			Trailing: newTopic,
		}, nil)

		return nil
	}

	if cnl.Topic() == "" {
		u.SendMessage(&message.Message{
			Prefix:  s.Config.ServerName,
			Command: message.RPL_NOTOPIC,
			Params: []string{
				u.NickName,
				channelName,
			},
			Trailing: "No topic is set.",
		})

		return nil
	} else {
		u.SendMessage(&message.Message{
			Prefix:  s.Config.ServerName,
			Command: message.RPL_TOPIC,
			Params: []string{
				u.NickName,
				channelName,
			},
			Trailing: cnl.Topic(),
		})

		// Sendout message RPL_TOPICWHOTIME

		return nil
	}

	return nil
}
