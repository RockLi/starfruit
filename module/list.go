/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"fmt"
	"github.com/flatpeach/ircd/channel"
	"github.com/flatpeach/ircd/message"
	"github.com/flatpeach/ircd/server"
	"github.com/flatpeach/ircd/user"
	"strings"
)

type List struct{}

func (module *List) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// LIST [ <channel> * ( "," <channel> ) ]

	var (
		channelsToList []*channel.Channel
	)

	if len(m.Params) == 0 {
		// List all channels
		channelsToList = s.GetAllChannels()
	} else {
		// List only specific channel
		channelNames := strings.Split(m.Params[0], ",")
		for _, channelName := range channelNames {
			cnl := s.FindChannelByName(channelName)
			if cnl != nil {
				channelsToList = append(channelsToList, cnl)
			}
		}
	}

	for _, cnl := range channelsToList {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.RPL_LIST,
			[]string{
				u.NickName,
				cnl.Name,
				fmt.Sprintf("%d", s.ChannelUserCount(cnl.Id)),
			},
			cnl.Topic(),
		))
	}

	u.SendMessage(message.New(
		s.Config.ServerName,
		message.RPL_LISTEND,
		[]string{u.NickName},
		"End of /LIST.",
	))

	return nil
}
