/*
 * Copyright 2014 The starfruit Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/server"
	"github.com/flatpeach/starfruit/user"
)

type Mode struct{}

func (module *Mode) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// MODE <nickname> *( ( "+" / "-" ) *( "i" / "w" / "o" / "O" / "r" ) )

	if len(m.Params) < 1 {
		u.SendErrorNeedMoreParams("MODE")
		return nil
	}

	nickName := m.Params[0]
	if u.NickName != nickName {
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.ERR_USERSDONTMATCH,
			[]string{
				u.NickName,
			},
			"Cannot change mode for other users",
		))

		return nil
	}

	if len(m.Params) == 1 {
		// Return current user modes
		u.SendMessage(message.New(
			s.Config.ServerName,
			message.RPL_UMODEIS,
			[]string{
				u.NickName,
				u.Modes(),
			},
			nil,
		))

		return nil
	}

	modes := m.Params[1]

	if modes[0] != '+' && modes[0] != '-' {
		return nil
	}

	operator := modes[0]

	for _, mode := range modes[1:] {

		if mode == 'a' {
			continue
		}

		if operator == '+' {
			if mode == 'o' || mode == 'O' {
				continue
			}
		}

		m := 0

		switch mode {
		case 'i':
			m = user.ModeInvisible
		case 'w':
			m = user.ModeReceiveWallops
		case 'r':
			m = user.ModeRestrictedUserConnection
		case 'o':
			m = user.ModeOperator
		case 'O':
			m = user.ModeLocalOperator
		case 's':
			m = user.ModeReceiveServiceNotice

		default:
			u.SendMessage(message.New(
				s.Config.ServerName,
				message.ERR_UMODEUNKNOWNFLAG,
				[]string{
					u.NickName,
				},
				"Unknown MODE flag",
			))

			return nil
		}

		if m != 0 {
			if operator == '+' {
				u.MarkMode(m)
			} else if operator == '-' {
				u.ClearMode(m)
			}
		}
	}

	u.SendMessage(message.New(
		s.Config.ServerName,
		"MODE",
		[]string{
			u.NickName,
		},
		modes,
	))

	return nil
}
