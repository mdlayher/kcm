//go:build linux
// +build linux

package kcm_test

import (
	"errors"
	"testing"

	"github.com/mdlayher/kcm"
	"golang.org/x/sys/unix"
)

func TestConn(t *testing.T) {
	c, err := kcm.Listen(kcm.Datagram, nil)
	if err != nil {
		// GitHub actions in February 2022 doesn't seem to support AF_KCM.
		if errors.Is(err, unix.EAFNOSUPPORT) {
			t.Skipf("skipping, AF_KCM not supported: %v", err)
		}

		t.Fatalf("failed to listen: %v", err)
	}
	defer c.Close()

	// TODO(mdlayher): finish.
}
