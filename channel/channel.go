/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package channel

import (
	"errors"
	"fmt"
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/user"
	"sync"
	"time"
)

// Channel Namespace/Prefix
const (
	NS_LOCAL_RAW               = "&"
	NS_NETWORK_RAW             = "#"
	NS_NETWORK_SAFE_RAW        = "!"
	NS_NETWORK_UNMODERATED_RAW = "+"
	NS_SOFTCHAN_RAW            = "."
	NS_GLOBAL_RAW              = "~"
)

const (
	NS_LOCAL               = 0x01 // '&'
	NS_NETWORK             = 0x02 // '#'
	NS_NETWORK_SAFE        = 0x04 // '!'
	NS_NETWORK_UNMODERATED = 0x08 // '+'
	NS_SOFTCHAN            = 0x10 // '.'
	NS_GLOBAL              = 0x20 // '~'
)

const (
	MODE_CREATOR    = 0x01
	MODE_OPERATOR   = 0x02
	MODE_VOICE      = 0x04
	MODE_ANONYMOUS  = 0x08
	MODE_INVITE     = 0x10
	MODE_MODERATED  = 0x20
	MODE_NO_MESSAGE = 0x40
	MODE_QUIET      = 0x80
	MODE_PRIVATE    = 0x100
	MODE_SECRET     = 0x200
	MODE_REOP       = 0x400
	MODE_KEY        = 0x800
	MODE_LIMIT      = 0x1000
	MODE_BAN        = 0x2000
	MODE_EXCEPTION  = 0x4000
	MODE_INVITATION = 0x8000
)

type Channel struct {
	Id        int
	Namespace int
	Name      string
	Modes     int

	topic        string
	users        []*user.User
	topicSetBy   string
	topicSettime int64
	mutex        sync.Mutex
}

const (
	MAX_NAME_LENGTH = 50
	MIN_NAME_LENGTH = 2
)

func New(s string) (*Channel, error) {
	if len(s) < MIN_NAME_LENGTH || len(s) > MAX_NAME_LENGTH {
		return nil, errors.New("Channel name too long or too short")
	}

	c := &Channel{
		users: make([]*user.User, 0),
	}

	switch s[0:1] {
	case NS_LOCAL_RAW:
		c.Namespace = NS_LOCAL

	case NS_NETWORK_RAW:
		c.Namespace = NS_NETWORK

	case NS_NETWORK_SAFE_RAW:
		c.Namespace = NS_NETWORK_SAFE

	case NS_NETWORK_UNMODERATED_RAW:
		c.Namespace = NS_NETWORK_UNMODERATED

	case NS_SOFTCHAN_RAW:
		c.Namespace = NS_SOFTCHAN

	case NS_GLOBAL_RAW:
		c.Namespace = NS_GLOBAL
	}

	c.Name = s[1:]

	return c, nil
}

func (c *Channel) String() string {
	var s = ""
	switch c.Namespace {
	case NS_LOCAL:
		s += NS_LOCAL_RAW

	case NS_NETWORK:
		s += NS_NETWORK_RAW

	case NS_NETWORK_SAFE:
		s += NS_NETWORK_SAFE_RAW

	case NS_NETWORK_UNMODERATED:
		s += NS_NETWORK_UNMODERATED_RAW

	case NS_SOFTCHAN:
		s += NS_SOFTCHAN_RAW

	case NS_GLOBAL:
		s += NS_GLOBAL_RAW
	}

	return s + c.Name
}

func (c *Channel) Join(newUser *user.User) error {

	for _, u := range c.users {
		if u.Id == newUser.Id {
			return errors.New("Duplicated user")
		}
	}

	c.users = append(c.users, newUser)

	return nil
}

func (c *Channel) Quit(uid int) error {
	for idx, u := range c.users {
		if u.Id == uid {
			users := c.users
			users[idx] = users[len(users)-1]
			users = users[:len(users)-1]
			c.users = users
			fmt.Println(c.users)
			return nil
		}
	}

	return errors.New("Failed to delete the user")
}

func (c *Channel) Broadcast(m *message.Message, exludes []int) error {
outer:
	for _, u := range c.users {

		if exludes != nil {
			for _, exludeId := range exludes {
				if exludeId == u.Id {
					continue outer
				}
			}
		}
		u.SendMessage(m)
	}

	return nil
}

func (c *Channel) Count() int {
	return len(c.users)
}

func (c *Channel) JoinedUsers() []*user.User {
	return c.users[:]
}

func (c *Channel) Exists(uid int) bool {
	for _, u := range c.users {
		if u.Id == uid {
			return true
		}
	}

	return false
}

func (c *Channel) SetTopic(s string, who string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.topic = s
	c.topicSetBy = who
	c.topicSettime = time.Now().Unix()
}

func (c *Channel) Topic() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.topic
}

func (c *Channel) TopicSetBy() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.topicSetBy
}

func (c *Channel) TopicSetTime() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.TopicSetTime()
}
