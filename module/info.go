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

type Info struct{}

func (module *Info) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// VERSION [ <target> ]

	u.SendMessage(message.New(
		s.Config.Server.Name,
		message.RPL_INFO,
		[]string{
			u.NickName,
		},
		"IRC",
	))

	u.SendMessage(message.New(
		s.Config.Server.Name,
		message.RPL_INFO,
		[]string{
			u.NickName,
		},
		"Released by starfruit.io written by Rock Lee and others contributors",
	))

	u.SendMessage(message.New(
		s.Config.Server.Name,
		message.RPL_INFO,
		[]string{
			u.NickName,
		},
		"Thank you for all of them, and contributing always welcome.",
	))

	u.SendMessage(message.New(
		s.Config.Server.Name,
		message.RPL_INFO,
		[]string{
			u.NickName,
		},
		"",
	))

	u.SendMessage(message.New(
		s.Config.Server.Name,
		message.RPL_INFO,
		[]string{
			u.NickName,
		},
		fmt.Sprintf("Started since %s", s.StartedAt.Format("Jan 2, 2006 at 3:04pm (MST)")),
	))

	u.SendMessage(message.New(
		s.Config.Server.Name,
		message.RPL_ENDOFINFO,
		[]string{
			u.NickName,
		},
		"End of /INFO list.",
	))

	return nil
}
