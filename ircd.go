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
	s        *server.Server
	commands map[string]interface{}
)

func doUserChecking() {
	for {
		time.Sleep(s.Config.PingUserInterval)

		ts := time.Now().Unix()

		log.Printf("[SERVER] Ready to scan the status of all users: %d", ts)

		users := s.GetAllUsers()
		for _, u := range users {
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

func doReply(u *user.User) {
	for m := range u.Out {
		data := m.String() + "\r\n"
		log.Printf("[Client:%s] Reply %s", u.Conn.RemoteAddr(), m.String())
		_, err := u.Conn.Write([]byte(data))
		if err != nil {
			log.Printf("[Client:%s] Failed to send reply message")
			// @Todo: send message to all joined channels of this user
			s.RemoveUser(u.Id)
		}
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
	go doReply(u)

	for {
		buf, _, err := reader.ReadLine()
		if err != nil {
			log.Printf("[Client:%s] Remote connection already closed!", u.Conn.RemoteAddr())
			close(u.In)
			close(u.Out)
			u.Conn.Close()
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
	s = server.New()
	s.Config = config.New()
	s.Config.Password = "1234567"
	s.Config.ServerName = "chat.starfruit.io"
	s.Config.SSL = false
	s.Config.BindIP = "127.0.0.1"
	s.Config.BindPort = 6667

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
	//registerCmd("NAMES", &module.Names{})
}

func main() {
	var (
		listener net.Listener = nil
		err      error
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
