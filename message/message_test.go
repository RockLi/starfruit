/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package message

import (
	"testing"
)

func TestMessage1(t *testing.T) {
	line := "USER rocklee 8 * :RockLee"

	m, err := Parse(line)

	if err != nil {
		t.Error("Should no error happened")
	}

	if m.Prefix != "" {
		t.Error("Prefix parsed error")
	}

	if m.Command != "USER" {
		t.Error("Command should be USER")
	}

	if m.Params[0] != "rocklee" {
		t.Error("Param error")
	}

	if m.Params[1] != "8" {
		t.Error("Param error")
	}

	if m.Params[2] != "*" {
		t.Error("Param error")
	}

	if m.Trailing != "RockLee" {
		t.Error("Trailing should be RockLee")
	}
}

func TestMessage2(t *testing.T) {
	line := ":Rock!rock@localhost PRIVMSG #abord :你好,世界!"

	m, err := Parse(line)
	if err != nil {
		t.Error("error happened")
	}

	if m.Prefix != "Rock!rock@localhost" {
		t.Error("Prefix parsed error")
	}

	if m.Command != "PRIVMSG" {
		t.Error("Command should be PRIVMSG")
	}

	if m.Params[0] != "#abord" {
		t.Error("Param parsed error")
	}

	if m.Trailing != "你好,世界!" {
		t.Error("Trailing parsed error")
	}

}

func TestMessageEmptyTrailing(t *testing.T) {
	line := ":prefix command param1 param2 :"

	m, err := Parse(line)

	if err != nil {
		t.Error("error happened")
	}

	if len(m.Params) != 3 {
		t.Error("should has param1, param2 and emptry trailing as the param3")
	}

	if !m.HasTrailing() || m.Trailing != "" {
		t.Error("message should has empty trailing part")
	}
}

func TestMessage3(t *testing.T) {
	line := ":prefix command param1 param2"

	m, err := Parse(line)

	if err != nil {
		t.Error("error happened")
	}

	if len(m.Params) != 2 {
		t.Error("should has param1, param2")
	}
}
