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
	_ "github.com/flatpeach/ircd/channel"
	"github.com/flatpeach/ircd/command"
	"github.com/flatpeach/ircd/config"
	"github.com/flatpeach/ircd/message"
	"github.com/flatpeach/ircd/module"
	"github.com/flatpeach/ircd/server"
	"github.com/flatpeach/ircd/user"
	"log"
	"net"
	_ "os/signal"
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
		for _, u := range users {

			if u.LastPongTime != 0 && ts-u.LastPongTime > int64(s.Config.UserTimeout) {
				// @Todo: Remove this use from server and sendout TimeOut message to users
				timeoutMsg := &message.Message{
					Prefix:  u.Full(),
					Command: "QUIT",
					Params: []string{
						fmt.Sprintf("ping timeout after %d seconds.", int64(s.Config.UserTimeout)),
					},
				}

				_ = timeoutMsg

				continue
			}

			u.SendMessage(&message.Message{
				Prefix:  s.Config.ServerName,
				Command: "PING",
				Params: []string{
					fmt.Sprintf("%d", ts),
				},
			})
		}

		log.Printf("[SERVER] Done to scan the status of all users for this time")

	}
}

func doRequest(u *user.User) {
	for buf := range u.In {
		m, err := message.NewFromRaw(string(buf))
		if err != nil {
			log.Printf("[Client:%s] Malformed message %s", u.Conn.RemoteAddr(), err)
			continue
		}

		log.Printf("[Client:%s] Request %s", u.Conn.RemoteAddr(), m)

		cmd, ok := commands[m.Command]
		if !ok {
			log.Printf("[Client:%s] Unknown command %s", u.Conn.RemoteAddr(), m.Command)
			continue
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

	for {
		buf, _, err := reader.ReadLine()
		if err != nil {
			log.Printf("[Client:%s] Remote connection already closed!", u.Conn.RemoteAddr())
			s.RemoveUser(u.Id)
			close(u.In)
			break
		}

		u.In <- buf

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
	s.Config = config.New()

	commands = make(map[string]interface{})

	registerCmd("PASS", &module.Pass{})
	registerCmd("USER", &module.User{})
	registerCmd("NICK", &module.Nick{})
	registerCmd("JOIN", &module.Join{})
	registerCmd("AWAY", &module.Away{})
	registerCmd("WHO", &module.Who{})
	registerCmd("PING", &module.Ping{})
	registerCmd("PONG", &module.Pong{})
	registerCmd("PART", &module.Part{})
	registerCmd("QUIT", &module.Quit{})
	registerCmd("LIST", &module.List{})
	registerCmd("WHOIS", &module.Whois{})
	registerCmd("PRIVMSG", &module.Privmsg{})
	registerCmd("ISON", &module.Ison{})
	//registerCmd("NAMES", &module.Names{})
}

func main() {
	flag.Parse()

	fmt.Println(configFile)

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
