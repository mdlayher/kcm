//go:build linux
// +build linux

package kcm_test

import (
	"testing"

	"github.com/mdlayher/kcm"
)

func TestConn(t *testing.T) {
	c, err := kcm.Listen(kcm.Datagram, nil)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer c.Close()

	// TODO(mdlayher): finish.
}
