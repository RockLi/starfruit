/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"fmt"
	"github.com/flatpeach/ircd/message"
	"github.com/flatpeach/ircd/server"
	"github.com/flatpeach/ircd/user"
	"log"
)

type Who struct{}

func (module *Who) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// WHO [ <mask> [ "o" ] ]

	if len(m.Params) == 0 && m.Trailing == "" {
		u.SendMessage(&message.Message{
			Prefix:   s.Config.ServerName,
			Command:  message.ERR_NEEDMOREPARAMS,
			Params:   nil,
			Trailing: "Need more params",
		})

		return nil
	}

	var (
		channelName string
		users       []*user.User
	)

	if m.Trailing != "" {
		channelName = m.Trailing
	} else {
		channelName = m.Params[0]
	}

	cnl := s.FindChannelByName(channelName)
	if cnl == nil {
		log.Printf("[COMMAND] WHO Malformed channel :%s", channelName)
		return nil
	}

	if !s.IsUserJoinedChannel(u.Id, cnl.Id) {
		log.Printf("User not joined!")
		goto endofwho
	}

	users = s.GetJoinedUsers(cnl.Id)
	for _, user := range users {
		u.SendMessage(&message.Message{
			Prefix:  s.Config.ServerName,
			Command: message.RPL_WHOREPLY,
			Params: []string{
				u.NickName,
				channelName,
				"~" + user.UserName,
				user.HostName,
				s.Config.ServerName,
				user.NickName,
				"H@",
			},
			Trailing: fmt.Sprintf("0 %s", user.RealName),
		})
	}

endofwho:
	u.SendMessage(&message.Message{
		Prefix:   s.Config.ServerName,
		Command:  message.RPL_ENDOFWHO,
		Params:   []string{u.NickName, channelName},
		Trailing: "End of /WHO list.",
	})

	return nil
}
