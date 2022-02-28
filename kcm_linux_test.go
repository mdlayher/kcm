//go:build linux
// +build linux

package kcm_test

import (
	"errors"
	"testing"

	"github.com/mdlayher/kcm"
	"golang.org/x/sys/unix"
)

func TestConnBasics(t *testing.T) {
	// This test doesn't need extra dependencies or privileges, so it only
	// covers a subset of the usual functionality.
	c, err := kcm.Listen(kcm.Datagram, nil)
	if err != nil {
		// GitHub actions in February 2022 doesn't seem to support AF_KCM.
		if errors.Is(err, unix.EAFNOSUPPORT) {
			t.Skipf("skipping, AF_KCM not supported: %v", err)
		}

		t.Fatalf("failed to listen: %v", err)
	}
	defer c.Close()

	// Clone the KCM socket so we have a maximum of N sockets.
	const n = 4
	cs := make([]*kcm.Conn, 0, n)
	cs = append(cs, c)

	for i := 0; i < n-1; i++ {
		cc, err := c.Clone()
		if err != nil {
			t.Fatalf("failed to clone KCM socket %d: %v", i, err)
		}
		defer cc.Close()

		cs = append(cs, cc)
	}

	// Since getsockname(2) is not supported, we can't really correlate these
	// sockets in any meaningful way. For now, just close everything out since
	// we don't want to perform privileged operations like loading eBPF
	// programs.
	for i, c := range cs {
		if err := c.Close(); err != nil {
			t.Fatalf("failed to close cs[%d]: %v", i, err)
		}
	}
}
