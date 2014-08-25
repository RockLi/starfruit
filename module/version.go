/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"fmt"
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/server"
	"github.com/flatpeach/starfruit/user"
	"github.com/flatpeach/starfruit/version"
)

type Version struct{}

func (module *Version) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// VERSION [ <target> ]

	u.SendMessage(message.New(
		s.Config.ServerName,
		message.RPL_VERSION,
		[]string{
			u.NickName,
			fmt.Sprintf("%d.%d(%s)", version.Major, version.Minor, version.PatchLevel),
			s.Config.ServerName,
		},
		version.MagicCode,
	))

	return nil
}
