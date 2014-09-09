/*
 * Copyright 2014 The starfruit Authors. All rights reserved.
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
	"github.com/flatpeach/starfruit/command"
	"github.com/flatpeach/starfruit/config"
	"github.com/flatpeach/starfruit/message"
	"github.com/flatpeach/starfruit/module"
	"github.com/flatpeach/starfruit/server"
	"github.com/flatpeach/starfruit/user"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	s                      *server.Server
	commands               map[string]interface{}
	configEnableAuth       bool
	configFile             string
	configServerName       string
	configPassword         string
	configIp               string
	configPorts            string
	configDisabledCommands string
	configMotdFile         string
	configSSL              bool
	configCertFile         string
	configKeyFile          string
	configPingUserInterval int
	configUserTimeout      int
)

func doUserChecking() {
	for {
		time.Sleep(time.Duration(s.Config.Recycle.PingInterval) * time.Second)
		ts := time.Now().Unix()
		log.Printf("[starfruit] Ready to scan the status of all users: %d", ts)

		users := s.GetAllUsers()
		for _, u := range users {
			if u.IsDisconnecting() {
				continue
			}

			if u.LastPongTime != 0 && ts-u.LastPongTime > int64(s.Config.Recycle.UserTimeout) {
				timeoutMsg := message.New(
					u.Full(),
					"QUIT",
					nil,
					fmt.Sprintf("ping timeout after %d seconds.", int64(s.Config.Recycle.UserTimeout)),
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
					fmt.Sprintf("Closing Link: %s (Ping timeout: %d seconds)", u.HostName, s.Config.Recycle.UserTimeout),
				))

				u.SendMessage(nil)
				u.EnterStatus(user.StatusDisconnecting)

				continue
			}

			u.SendMessage(message.New(
				s.Config.Server.Name,
				"PING",
				[]string{
					fmt.Sprintf("%d", ts),
				},
				nil,
			))
		}

		log.Printf("[starfruit] Done to scan the status of all users for this time")

	}
}

func doResponse(u *user.User) {
	for buf := range u.Out {
		if buf == nil {
			u.Close()
			return
		}
		log.Printf("[Client:%s] Reply %s", u.Conn.RemoteAddr(), string(buf))
		_, err := u.Conn.Write(buf)
		if err != nil {
			log.Printf("[Client:%s] Failed to send reply message")
			u.SendMessage(nil)
		}
	}
}

func doRequest(u *user.User) {
	for buf := range u.In {

		if buf == nil && u.IsDisconnecting() {
			return
		}

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
					u.Config.Server.Name,
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
					u.Config.Server.Name,
					message.ERR_NOTREGISTERED,
					[]string{"*"},
					"You have not registered",
				))

				continue
			}
		} else {
			if m.Command == "PASS" || m.Command == "USER" || m.Command == "SERVICE" {
				u.SendMessage(message.New(
					u.Config.Server.Name,
					message.ERR_ALREADYREGISTRED,
					[]string{u.NickName},
					"Already registered",
				))

				continue
			}
		}

		if len(s.Config.Server.DisabledCommands) > 0 {
			for _, command := range s.Config.Server.DisabledCommands {
				if command == m.Command {
					switch m.Command {
					case "USERS":
						u.SendMessage(message.New(
							u.Config.Server.Name,
							message.ERR_USERSDISABLED,
							[]string{u.NickName},
							"USERS has been disabled",
						))

					}

					continue
				}
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
			u.SendMessage(nil)
			u.EnterStatus(user.StatusDisconnecting)
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

func doListen(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[starfruit] Accept error: %s", err)
			break
		}

		log.Printf("[starfruit] Accepted connection from: %s", conn.RemoteAddr())

		u := user.New(s.Config, conn)

		go doConn(u)
	}

}

func init() {
	flag.StringVar(&configFile, "config", "", "config file of this irc server")
	flag.BoolVar(&configEnableAuth, "enable-auth", false, "enable auth or not")
	flag.StringVar(&configServerName, "name", "", "name of this IRC server")
	flag.StringVar(&configPassword, "password", "", "password to join this server")
	flag.StringVar(&configIp, "ip", "", "ip to bind, normally to choose 0.0.0.0")
	flag.StringVar(&configPorts, "ports", "", "ports to bind, for multiple ports use comma to separate them")
	flag.StringVar(&configDisabledCommands, "disabled-commands", "", "commands to disable of this server")
	flag.StringVar(&configMotdFile, "motd", "", "motd file")
	flag.BoolVar(&configSSL, "enable-ssl", false, "whether to enable ssl")
	flag.StringVar(&configCertFile, "cert-file", "", "")
	flag.StringVar(&configKeyFile, "key-file", "", "")
	flag.IntVar(&configPingUserInterval, "ping-interval", -1, "")
	flag.IntVar(&configUserTimeout, "user-timeout", -1, "")

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
	registerCmd("MODE", &module.Mode{})
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
	registerCmd("USERS", &module.Users{})
	registerCmd("VERSION", &module.Version{})
	registerCmd("WHO", &module.Who{})
	registerCmd("WHOIS", &module.Whois{})
}

func main() {
	var (
		listener net.Listener = nil
		err      error
	)

	flag.Parse()

	if configFile != "" {
		err := s.Config.LoadFromFile(configFile)
		if err != nil {
			log.Fatalf("[starfruit] Failed to load the configuration file :%s", err)
			return
		}
		log.Printf("[starfruit] Load the configuration file :%s", configFile)
	}

	/* Handle overrided parameters from command line */
	if configIp != "" {
		s.Config.Server.Ip = configIp
	}

	if configPorts != "" {
		ports := strings.Split(configPorts, ",")
		for _, port := range ports {
			s.Config.Server.Ports = nil
			port, err := strconv.Atoi(port)
			if err != nil {
				log.Fatalf("[starfruit] Port specified error, %s", err)
				return
			}
			s.Config.Server.Ports = append(s.Config.Server.Ports, port)
		}
	}

	if configServerName != "" {
		s.Config.Server.Name = configServerName
	}

	if configEnableAuth && configPassword != "" {
		s.Config.Server.Password = configPassword
	}

	if configMotdFile != "" {
		s.Config.Motd.File = configMotdFile
	}

	if configSSL {
		s.Config.Server.SSL = configSSL
	}

	if configCertFile != "" {
		s.Config.Server.CertFile = configCertFile
	}

	if configKeyFile != "" {
		s.Config.Server.KeyFile = configKeyFile
	}

	if configPingUserInterval > -1 {
		s.Config.Recycle.PingInterval = configPingUserInterval
	}

	if configUserTimeout > -1 {
		s.Config.Recycle.UserTimeout = configUserTimeout
	}

	if configDisabledCommands != "" {
		commands := strings.Split(configDisabledCommands, ",")
		s.Config.Server.DisabledCommands = commands
	}

	/* Listen on all ports */
	for _, port := range s.Config.Server.Ports {
		if s.Config.Server.SSL {
			cert, err := tls.LoadX509KeyPair(s.Config.Server.CertFile, s.Config.Server.KeyFile)
			if err != nil {
				log.Fatal("[starfruit] Failed to load certificates!")
				return
			}

			config := tls.Config{Certificates: []tls.Certificate{cert}}
			config.Rand = rand.Reader

			listener, err = tls.Listen("tcp", fmt.Sprintf("%s:%d", s.Config.Server.Ip, port), &config)
			if err != nil {
				log.Fatalf("[starfruit] Failed to start the SERVER(SSL), %s", err)
				return
			}
		} else {
			listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.Config.Server.Ip, port))
			if err != nil {
				log.Fatalf("[starfruit] Failed to start the SERVER, %s", err)
			}
		}

		go doListen(listener)
	}

	log.Printf("[starfruit] Server started at %s",
		fmt.Sprintf("%s[%v]", s.Config.Server.Ip, s.Config.Server.Ports))

	go doUserChecking()

	for {
		time.Sleep(10000)
	}

}
