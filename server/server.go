/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package server

import (
	"fmt"
	"github.com/flatpeach/ircd/channel"
	"github.com/flatpeach/ircd/config"
	"github.com/flatpeach/ircd/message"
	"github.com/flatpeach/ircd/user"
	"sync"
)

type Server struct {
	Config *config.Config // Config for current IRC Server

	channels map[int]*channel.Channel // All channels in this server
	users    map[int]*user.User       // All users existed in this server

	userToChannels map[int][]int // User to channels list

	maxUserId    int // Current the max user id
	maxChannelId int // current the max channel id

	mutex sync.Mutex
}

func New() *Server {
	s := &Server{
		Config: nil,

		channels:       make(map[int]*channel.Channel),
		users:          make(map[int]*user.User),
		userToChannels: make(map[int][]int),

		maxUserId:    0,
		maxChannelId: 0,
	}

	return s
}

func (s *Server) GetAllChannels() []*channel.Channel {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var channels []*channel.Channel

	for _, cnl := range s.channels {
		channels = append(channels, cnl)
	}

	return channels
}

func (s *Server) GetAllUsers() []*user.User {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var users []*user.User

	for _, u := range s.users {
		users = append(users, u)
	}

	return users
}

func (s *Server) ChannelUserCount(cid int) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cnl := s.channels[cid]
	if cnl != nil {
		return cnl.Count()
	}

	return 0
}

func (s *Server) FindChannelByName(c string) *channel.Channel {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, v := range s.channels {
		if v.String() == c {
			return v
		}
	}

	return nil
}

func (s *Server) FindChannelById(cid int) *channel.Channel {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	c, exists := s.channels[cid]
	if exists {
		return c
	}

	return nil
}

func (s *Server) CreateChannel(name string) (*channel.Channel, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.createChannel(name)
}

func (s *Server) createChannel(name string) (*channel.Channel, error) {
	c, err := channel.New(name)
	if err != nil {
		return nil, err
	}

	s.maxChannelId += 1
	c.Id = s.maxChannelId
	s.channels[c.Id] = c

	return c, nil
}

func (s *Server) FindOrCreateChannel(name string) (*channel.Channel, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, v := range s.channels {
		if v.String() == name {
			return v, nil
		}
	}

	return s.createChannel(name)
}

func (s *Server) RemoveChannel(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	c := s.FindChannelByName(name)
	if c != nil {
		delete(s.channels, c.Id)
	}

	return nil
}

func (s *Server) MaxUserId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.maxUserId
}

func (s *Server) NewUserId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.maxUserId += 1
	return s.maxUserId
}

func (s *Server) MaxChannelId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.maxChannelId
}

func (s *Server) NewChannelId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.maxChannelId += 1
	return s.maxChannelId
}

func (s *Server) RegisterUser(u *user.User) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.users[u.Id] = u
	return true, nil
}

func (s *Server) GetUserByNickName(name string) *user.User {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, user := range s.users {
		if user.NickName == name {
			return user
		}
	}

	return nil
}

func (s *Server) RemoveUser(uid int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	channels := s.getJoinedChannels(uid)

	fmt.Println(channels)

	for _, cnl := range channels {
		cnl.Quit(uid)
	}

	delete(s.users, uid)
	delete(s.userToChannels, uid)

}

func (s *Server) ExistsUser(uid int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.users[uid]
	return exists
}

func (s *Server) IsUserJoinedChannel(uid int, cid int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cnl := s.channels[cid]
	if cnl != nil {
		return cnl.Exists(uid)
	}

	return false
}

func (s *Server) getJoinedChannels(uid int) []*channel.Channel {
	cids, exists := s.userToChannels[uid]
	if !exists {
		return nil
	}

	var channels []*channel.Channel

	for _, cid := range cids {
		channels = append(channels, s.channels[cid])
	}

	return channels
}

func (s *Server) GetJoinedChannels(uid int) []*channel.Channel {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.getJoinedChannels(uid)
}

func (s *Server) GetJoinedUsers(cid int) []*user.User {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cnl := s.channels[cid]
	if cnl != nil {
		return cnl.JoinedUsers()
	}

	return nil
}

func (s *Server) BroadcastMessage(cid int, m *message.Message, excludeIds []int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cnl := s.channels[cid]
	if cnl != nil {
		cnl.Broadcast(m, excludeIds)
	}
}

func (s *Server) JoinChannel(uid int, cid int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cnl := s.channels[cid]
	if cnl != nil {
		u := s.users[uid]
		if u != nil {
			cnl.Join(u)
		}
	}

	cids := s.userToChannels[uid]
	if cids != nil {
		for _, channelId := range cids {
			if channelId == cid {
				// Already in this channel
				return
			}
		}

	}
	s.userToChannels[uid] = append(s.userToChannels[uid], cid)
}

func (s *Server) QuitFromChannel(uid int, cid int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cnl := s.channels[cid]
	if cnl != nil {
		cnl.Quit(uid)
	}
}
