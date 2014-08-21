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
	"log"
	"strings"
)

type Join struct{}

func (module *Join) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// JOIN (<channel> ) / "0"

	if len(m.Params) == 1 && m.Params[0] == "0" {
		// This user wanna leave all channels he/she joined now
		// Send PART replies to members of each channel
		partMsg := &message.Message{}

		joinedChannels := s.GetJoinedChannels(u.Id)
		for _, channel := range joinedChannels {
			s.BroadcastMessage(channel.Id, partMsg, nil)
		}

		return nil
	}

	channelsRaw := m.Params[0]
	// @Todo: handle channel join key

	channels := strings.Split(channelsRaw, ",")

	for _, channelRaw := range channels {
		cnl, err := s.FindOrCreateChannel(channelRaw)
		if err != nil {
			log.Printf("[JOIN] Malformed channel :%s", channelRaw)
			return nil
		}

		joinMsg := &message.Message{
			Prefix:   u.Full(),
			Command:  "JOIN",
			Params:   nil,
			Trailing: cnl.String(),
		}

		u.SendMessage(joinMsg)

		u.SendMessage(&message.Message{
			Prefix:  s.Config.ServerName,
			Command: "MODE",
			Params:  []string{cnl.String(), "+nt"},
		})

		u.SendMessage(&message.Message{
			Prefix:  s.Config.ServerName,
			Command: message.RPL_NAMREPLY,
			Params: []string{
				u.NickName,
				"=",
				cnl.String(),
			},
			Trailing: (func() string {
				var names []string = []string{"@" + u.NickName}
				users := s.GetJoinedUsers(cnl.Id)
				for _, u := range users {
					names = append(names, "@"+u.NickName)
				}
				return strings.Join(names, " ")
			})(),
		})

		u.SendMessage(&message.Message{
			Prefix:  s.Config.ServerName,
			Command: message.RPL_ENDOFNAMES,
			Params: []string{
				u.NickName,
				cnl.String(),
			},
			Trailing: "End of /NAMES list.",
		})

		cnl.Broadcast(joinMsg, nil)
		s.JoinChannel(u.Id, cnl.Id)

		// @Todo: Fix duplicated created channels in client side

	}

	return nil
}
