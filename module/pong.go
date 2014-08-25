/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package module

import (
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/server"
	"github.com/flatpeach/starfruit/user"
	"time"
)

type Pong struct{}

func (module *Pong) Handle(s *server.Server, u *user.User, m *message.Message) error {
	// PONG

	// @Todo: fullfil this part

	ts := time.Now().Unix()
	u.LastPongTime = ts

	return nil
}
