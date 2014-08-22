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

	if len(m.Params) == 0 {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_NEEDMOREPARAMS,
			nil,
			"Need more params",
		))

		return nil
	}

	var (
		channelName string
		users       []*user.User
	)

	channelName = m.Params[0]

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
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.RPL_WHOREPLY,
			[]string{
				u.NickName,
				channelName,
				"~" + user.UserName,
				user.HostName,
				s.Config.ServerName,
				user.NickName,
				"H@",
			},
			fmt.Sprintf("0 %s", user.RealName),
		))
	}

endofwho:
	u.SendMessage(message.New(
		s.Config.ServerName,
		message.RPL_ENDOFWHO,
		[]string{u.NickName, channelName},
		"End of /WHO list.",
	))

	return nil
}
