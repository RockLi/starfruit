/*
 * Copyright 2014 The starfruit Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"fmt"
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/server"
	"github.com/flatpeach/starfruit/user"
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
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NOSUCHCHANNEL,
			[]string{u.NickName, channelName},
			"No such channel",
		))

		return nil
	}

	if !s.IsUserJoinedChannel(u.Id, cnl.Id) {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NOTONCHANNEL,
			[]string{u.NickName, channelName},
			"You're not on that channel",
		))

		return nil
	}

	if len(m.Params) > 1 {
		var newTopic = m.Params[1]
		cnl.SetTopic(newTopic, u.Full())

		s.BroadcastMessage(cnl.Id, message.New(
			u.Full(),
			"TOPIC",
			[]string{
				channelName,
			},
			newTopic,
		), nil)

		return nil
	}

	if cnl.Topic() == "" {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.RPL_NOTOPIC,
			[]string{
				u.NickName,
				channelName,
			},
			"No topic is set.",
		))

		return nil
	} else {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.RPL_TOPIC,
			[]string{
				u.NickName,
				channelName,
			},
			cnl.Topic(),
		))

		u.SendMessage(message.New(
			s.Config.ServerName,
			message.RPL_TOPICWHOTIME,
			[]string{
				u.NickName,
				channelName,
				cnl.TopicSetBy(),
				fmt.Sprintf("%d", cnl.TopicSetTime()),
			},
			nil,
		))

		// Sendout message RPL_TOPICWHOTIME
	}

	return nil
}
