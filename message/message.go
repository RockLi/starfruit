/*
 * Copyright 2014 The starfruit Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package message

import (
	"errors"
	"strings"
)

// Message Format
// :<prefix> <command> <params> :<trailing>

type Message struct {
	Prefix   string
	Command  string
	Params   []string
	Trailing string

	hasPrefix   bool
	hasTrailing bool // Especially useful if the trailing part is present but with empty string
}

func Parse(s string) (*Message, error) {
	var (
		m             *Message
		err           error = nil
		prefixEnd     int   = -1
		trailingStart int   = len(s)
		arr           []string
	)

	m = &Message{}

	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, ":") {
		prefixEnd = strings.Index(s, " ")
		if prefixEnd-1 <= 0 {
			err = errors.New("Malformed Prefix Message")
			goto ret
		}

		m.Prefix = s[1:prefixEnd]
		m.hasPrefix = true
	}

	trailingStart = strings.Index(s, " :")
	if trailingStart >= 0 {
		m.Trailing = s[trailingStart+2:]
		m.Params = append(m.Params, m.Trailing)
		m.hasTrailing = true
	} else {
		trailingStart = len(s)
	}

	arr = strings.Split(s[prefixEnd+1:trailingStart], " ")
	if len(arr) == 0 {
		err = errors.New("Malformed Message")
		goto ret
	}

	m.Command = arr[0]
	m.Params = append(arr[1:], m.Params[:]...)

ret:
	return m, err
}

func New(prefix interface{}, command string, params []string, trailing interface{}) *Message {
	m := &Message{
		Command: command,
		Params:  params,
	}

	if prefix != nil {
		m.Prefix = prefix.(string)
		m.hasPrefix = true
	} else {
		m.Prefix = ""
		m.hasPrefix = false
	}

	if trailing != nil {
		m.SetHasTrailing(true)
		m.Trailing = trailing.(string)
		m.Params = append(m.Params, m.Trailing)
	} else {
		m.SetHasTrailing(false)
		m.Trailing = ""
	}

	return m

}

func (m *Message) String() string {
	var s string

	if m.Prefix != "" {
		s += ":"
		s += m.Prefix
	}

	if m.Command != "" {
		if s != "" {
			s += " "
		}
		s += m.Command
	}

	if m.Params != nil &&
		((!m.hasTrailing && len(m.Params) > 0) ||
			m.hasTrailing && len(m.Params) > 1) {
		if s != "" {
			s += " "
		}

		if m.hasTrailing {
			s += strings.Join(m.Params[:len(m.Params)-1], " ")
		} else {
			s += strings.Join(m.Params, " ")
		}
	}

	if m.hasTrailing {
		if s != "" {
			s += " :"
		}
		s += m.Trailing
	}

	return s
}

func (m *Message) HasTrailing() bool {
	return m.hasTrailing
}

func (m *Message) SetHasTrailing(b bool) {
	m.hasTrailing = b
}

func (m *Message) HasPrefix() bool {
	return m.hasPrefix
}

func (m *Message) SetHasPrefix(b bool) {
	m.hasPrefix = b
}
