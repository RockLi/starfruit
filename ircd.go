/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/flatpeach/ircd/command"
	"github.com/flatpeach/ircd/config"
	"github.com/flatpeach/ircd/message"
	"github.com/flatpeach/ircd/module"
	"github.com/flatpeach/ircd/server"
	"github.com/flatpeach/ircd/user"
	"log"
	"net"
	"time"
)

var (
	s          *server.Server
	commands   map[string]interface{}
	configFile string
)

func doUserChecking() {
	for {
		time.Sleep(time.Duration(s.Config.PingUserInterval) * time.Second)
		ts := time.Now().Unix()
		log.Printf("[SERVER] Ready to scan the status of all users: %d", ts)

		users := s.GetAllUsers()
		fmt.Printf("users: %v\n", users)
		for _, u := range users {
			if u.IsDisconnecting() {
				continue
			}

			if u.LastPongTime != 0 && ts-u.LastPongTime > int64(s.Config.UserTimeout) {
				timeoutMsg := message.New(
					u.Full(),
					"QUIT",
					nil,
					fmt.Sprintf("ping timeout after %d seconds.", int64(s.Config.UserTimeout)),
				)

				s.RemoveUser(u.Id)

				channels := s.GetJoinedChannels(u.Id)
				for _, cnl := range channels {
					s.BroadcastMessage(cnl.Id, timeoutMsg, nil)
				}

				u.SendMessage(message.New(
					nil,
					"ERROR",
					nil,
					fmt.Sprintf("Closing Link: %s (Ping timeout: %d seconds)", u.HostName, s.Config.UserTimeout),
				))

				continue
			}

			u.SendMessage(message.New(
				s.Config.ServerName,
				"PING",
				[]string{
					fmt.Sprintf("%d", ts),
				},
				nil,
			))
		}

		log.Printf("[SERVER] Done to scan the status of all users for this time")

	}
}

func doResponse(u *user.User) {
	for buf := range u.Out {
		log.Printf("[Client:%s] Reply %s", u.Conn.RemoteAddr(), string(buf))
		_, err := u.Conn.Write(buf)
		if err != nil {
			log.Printf("[Client:%s] Failed to send reply message")
			u.Close()
		}
	}
}

func doRequest(u *user.User) {
	for buf := range u.In {
		m, err := message.Parse(string(buf))
		if err != nil {
			log.Printf("[Client:%s] Malformed message %s", u.Conn.RemoteAddr(), err)
			continue
		}

		log.Printf("[Client:%s] Request %s", u.Conn.RemoteAddr(), m)

		cmd, ok := commands[m.Command]
		if !ok {
			log.Printf("[Client:%s] Unknown command %s", u.Conn.RemoteAddr(), m.Command)
			if u.IsRegistered() {
				u.SendMessage(message.New(
					u.Config.ServerName,
					message.ERR_UNKNOWNCOMMAND,
					[]string{u.NickName, m.Command},
					"Unknown command",
				))
			}

			continue
		}

		if !u.IsRegistered() {
			// We only allow limited commands before user registered successfully
			if m.Command != "PASS" && m.Command != "USER" && m.Command != "NICK" {
				u.SendMessage(message.New(
					u.Config.ServerName,
					message.ERR_NOTREGISTERED,
					[]string{"*"},
					"You have not registered",
				))

				continue
			}
		} else {
			if m.Command == "PASS" || m.Command == "USER" || m.Command == "SERVICE" {
				u.SendMessage(message.New(
					u.Config.ServerName,
					message.ERR_ALREADYREGISTRED,
					[]string{u.NickName},
					"Already registered",
				))

				continue
			}
		}

		err = cmd.(command.Command).Handle(s, u, m)

		if err != nil {
			log.Printf("Client:%s] Error %s", err)
			continue
		}
	}
}

func doConn(u *user.User) {
	reader := bufio.NewReader(u.Conn)

	go doRequest(u)
	go doResponse(u)

	for {
		buf, _, err := reader.ReadLine()
		if err != nil {
			log.Printf("[Client:%s] Remote connection already closed!", u.Conn.RemoteAddr())
			s.RemoveUser(u.Id)
			u.Close()
			break
		}

		if len(buf) > 0 {
			u.In <- buf
		}
	}
}

func registerCmd(cmd string, v interface{}) {
	_, ok := commands[cmd]
	if ok {
		log.Fatalf("[SERVER] Command %s already registered!", cmd)
		return
	}

	commands[cmd] = v
}

func init() {
	flag.StringVar(&configFile, "config", "./ircd.conf", "config file of this irc server")

	s = server.New()
	s.StartedAt = time.Now()
	s.Config = config.New()

	commands = make(map[string]interface{})

	registerCmd("AWAY", &module.Away{})
	registerCmd("INFO", &module.Info{})
	registerCmd("INVITE", &module.Invite{})
	registerCmd("ISON", &module.Ison{})
	registerCmd("JOIN", &module.Join{})
	registerCmd("LIST", &module.List{})
	registerCmd("MOTD", &module.Motd{})
	registerCmd("NICK", &module.Nick{})
	registerCmd("PART", &module.Part{})
	registerCmd("PASS", &module.Pass{})
	registerCmd("PING", &module.Ping{})
	registerCmd("PONG", &module.Pong{})
	registerCmd("PRIVMSG", &module.Privmsg{})
	registerCmd("QUIT", &module.Quit{})
	registerCmd("TIME", &module.Time{})
	registerCmd("TOPIC", &module.Topic{})
	registerCmd("USER", &module.User{})
	registerCmd("VERSION", &module.Version{})
	registerCmd("WHO", &module.Who{})
	registerCmd("WHOIS", &module.Whois{})

}

func main() {
	flag.Parse()
	err := s.Config.LoadFromJSONFile(configFile)
	if err != nil {
		log.Fatalf("[SERVER] Failed to load the configuration file :%s", err)
		return
	}

	log.Printf("[SERVER] Load the configuration file :%s", configFile)

	var (
		listener net.Listener = nil
	)

	if s.Config.SSL {
		cert, err := tls.LoadX509KeyPair("./certs/cert.pem", "./certs/key.pem")
		if err != nil {
			log.Fatal("[SERVER] Failed to load certificates!")
			return
		}

		config := tls.Config{Certificates: []tls.Certificate{cert}}
		config.Rand = rand.Reader

		listener, err = tls.Listen("tcp", fmt.Sprintf("%s:%d", s.Config.BindIP, s.Config.BindPort), &config)
		if err != nil {
			log.Fatal("[SERVER] Failed to start the SERVER(SSL)")
			return
		}

	} else {

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.Config.BindIP, s.Config.BindPort))
		if err != nil {
			log.Fatal("[SERVER] Failed to start the SERVER")
		}

	}

	log.Printf("[SERVER] Server started at %s", fmt.Sprintf("%s:%d", s.Config.BindIP, s.Config.BindPort))

	log.Printf("[SERVER] Server started Goroutine doReply")

	go doUserChecking()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[SERVER] Accept error: %s", err)
			break
		}

		log.Printf("[SERVER] Accepted connection from: %s", conn.RemoteAddr())

		u := user.New(s.Config, conn)

		go doConn(u)
	}

}
