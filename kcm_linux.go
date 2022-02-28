//go:build linux
// +build linux

package kcm

import (
	"errors"
	"syscall"

	"github.com/mdlayher/socket"
	"golang.org/x/sys/unix"
)

// A conn backs Conn for AF_KCM sockets. We can use socket.Conn directly on
// Linux to implement most of the necessary methods.
type conn = socket.Conn

// listen is the entry point for Listen on Linux.
func listen(styp Type, _ *Config) (*Conn, error) {
	// Convert Type to the matching SOCK_* constant.
	var typ int
	switch styp {
	case Datagram:
		typ = unix.SOCK_DGRAM
	case SequencedPacket:
		typ = unix.SOCK_SEQPACKET
	default:
		return nil, errors.New("kcm: invalid Type value")
	}

	// KCMPROTO_CONNECTED is the only valid proto value.
	c, err := socket.Socket(unix.AF_KCM, typ, unix.KCMPROTO_CONNECTED, "kcm", nil)
	if err != nil {
		return nil, err
	}

	return &Conn{c: c}, nil
}

// clone calls ioctl(2) to clone Conn.
func (c *Conn) clone() (*Conn, error) {
	cc, err := c.c.IoctlKCMClone()
	if err != nil {
		return nil, err
	}

	return &Conn{c: cc}, nil
}

// attach calls ioctl(2) to attach a raw client connection and eBPF file
// descriptor to Conn's multiplexor.
func (c *Conn) attach(rc syscall.RawConn, bpfFD int) error {
	var err error
	doErr := rc.Control(func(fd uintptr) {
		err = c.c.IoctlKCMAttach(unix.KCMAttach{
			Fd:     int32(fd),
			Bpf_fd: int32(bpfFD),
		})
	})
	if doErr != nil {
		return doErr
	}

	return err
}

// wait waits for ClientConn completion and calls ioctl(2) to unattach the
// client connection from Conn's multiplexor.
func (cc *ClientConn) wait() error {
	// We own this client connection, clean it up on return.
	defer tryClose(cc.sc)

	// Read SO_ERROR to determine whether or not a client has disconnected.
	//
	// TODO(mdlayher): it seems we don't always get notified on client
	// disconnect, especially in concurrent scenarios. Investigate.
	var err error
	doErr := cc.rc.Read(func(fd uintptr) bool {
		switch serr := soError(fd); serr {
		case unix.EPIPE:
			// Kernel woke us due to client error, unattach from KCM.
			err = cc.c.c.IoctlKCMUnattach(unix.KCMUnattach{Fd: int32(fd)})
			return true
		case nil:
			// Waiting for readiness.
			return false
		default:
			// Terminate and unattach due to unknown error.
			err = serr
			_ = cc.c.c.IoctlKCMUnattach(unix.KCMUnattach{Fd: int32(fd)})
			return true
		}
	})
	if doErr != nil {
		return doErr
	}

	return err
}

// soError calls getsockopt(2) for SO_ERROR.
func soError(fd uintptr) error {
	errno, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_ERROR)
	if err != nil {
		return err
	}
	if errno == 0 {
		return nil
	}

	return unix.Errno(errno)
}
