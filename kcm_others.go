//go:build !linux
// +build !linux

package kcm

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
)

// errUnimplemented is returned by all functions on non-Linux platforms.
var errUnimplemented = fmt.Errorf("kcm: not implemented on %s", runtime.GOOS)

func listen(_ Type, _ *Config) (*Conn, error) { return nil, errUnimplemented }

type conn struct{}

func (*Conn) clone() (*Conn, error)                 { return nil, errUnimplemented }
func (*Conn) attach(_ syscall.RawConn, _ int) error { return errUnimplemented }

func (*ClientConn) wait() error { return errUnimplemented }

func (*conn) Close() error                       { return errUnimplemented }
func (*conn) Read(_ []byte) (int, error)         { return 0, errUnimplemented }
func (*conn) Write(_ []byte) (int, error)        { return 0, errUnimplemented }
func (*conn) SetDeadline(_ time.Time) error      { return errUnimplemented }
func (*conn) SetReadDeadline(_ time.Time) error  { return errUnimplemented }
func (*conn) SetWriteDeadline(_ time.Time) error { return errUnimplemented }
