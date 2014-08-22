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

type Motd struct{}

func (module *Motd) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// MOTD <target>

	u.SendMotd()

	return nil
}
