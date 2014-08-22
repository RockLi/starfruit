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
)

type Pass struct{}

func (module *Pass) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// PASS param
	var pwd string

	if len(m.Params) > 0 {
		pwd = m.Params[0]
	}

	if s.Config.Password == "" {
		log.Printf("[COMMAND] PASS :no need to verify your password, server disabled that")
		return nil
	}

	if pwd != s.Config.Password {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_PASSWDMISMATCH,
			[]string{"*"},
			"Password incorrect",
		))

		return nil
	}

	u.EnterStatus(user.PasswordVerified)

	return nil
}
