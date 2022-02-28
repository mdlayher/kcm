package kcm

import (
	"io"
	"syscall"
	"time"
)

// TODO(mdlayher): Conn is not quite a net.Conn or net.PacketConn and it may not
// make sense to treat it as such. Do some research to determine if trying to
// fully comply with either of these interfaces is reasonable. In the meantime,
// io.ReadWriteCloser is a decent compromise.

var _ io.ReadWriteCloser = &Conn{}

// A Conn is a Linux Kernel Connection Multiplexor (AF_KCM) connection.
type Conn struct {
	c *conn
}

// Type is a socket type used when creating a Conn with Listen.
//enumcheck:exhaustive
type Type int

// Possible Type values. Note that the zero value is not valid: callers must
// always specify one of Datagram or SequencedPacket when calling Listen.
const (
	_ Type = iota
	Datagram
	SequencedPacket
)

// Config contains options for a Conn.
type Config struct{}

// Listen opens a Kernel Connection Multiplexor socket using the given socket
// type.
//
// The socket type must be one of the Type constants: Datagram or
// SequencedPacket.
//
// The Config specifies optional configuration for the Conn. A nil *Config
// applies the default configuration.
func Listen(typ Type, cfg *Config) (*Conn, error) {
	c, err := listen(typ, cfg)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Close implements io.Closer.
func (c *Conn) Close() error { return c.c.Close() }

// Clone produces a cloned Conn which is attached to the same kernel multiplexor
// as the existing Conn. Connections will be distributed by the kernel to any
// Conn attached to the same multiplexor.
func (c *Conn) Clone() (*Conn, error) { return c.clone() }

// Read implements io.Reader.
func (c *Conn) Read(b []byte) (int, error) { return c.c.Read(b) }

// Write implements io.Writer.
func (c *Conn) Write(b []byte) (int, error) { return c.c.Write(b) }

// SetDeadline sets a read and write deadline for Conn.
func (c *Conn) SetDeadline(t time.Time) error { return c.c.SetDeadline(t) }

// SetReadDeadline sets a read deadline for Conn.
func (c *Conn) SetReadDeadline(t time.Time) error { return c.c.SetReadDeadline(t) }

// SetWriteDeadline sets a write deadline for Conn.
func (c *Conn) SetWriteDeadline(t time.Time) error { return c.c.SetWriteDeadline(t) }

// A ClientConn is a client TCP connection which has been attached to a KCM Conn
// multiplexor. A ClientConn is created using the Conn.Attach method.
type ClientConn struct {
	// Server AF_KCM socket.
	c *Conn

	// Client listener accepted AF_INET{,6} socket.
	sc syscall.Conn
	rc syscall.RawConn
}

// Attach attaches a given client TCP connection to the multiplexor attached to
// Conn, using the input eBPF program file descriptor to frame incoming network
// protocol bytes into atomic messages.
//
// The ClientConn produced by Attach takes ownership of the TCP connection; no
// further methods (including Close) should be called on that connection. See
// ClientConn.Wait to clean up a completed client connection.
func (c *Conn) Attach(sc syscall.Conn, bpfFD int) (*ClientConn, error) {
	// We own sc: make sure to close it on error.
	rc, err := sc.SyscallConn()
	if err != nil {
		_ = tryClose(sc)
		return nil, err
	}

	if err := c.attach(rc, bpfFD); err != nil {
		_ = tryClose(sc)
		return nil, err
	}

	return &ClientConn{
		c:  c,
		sc: sc,
		rc: rc,
	}, nil
}

// Wait waits for a client TCP connection to terminate, then unattaches it from
// the underlying Conn's multiplexor. Wait should be used to clean up resources
// from completed client connections.
func (cc *ClientConn) Wait() error { return cc.wait() }

// tryClose attempts to call Close on the input syscall.Conn, since effectively
// every implementation of syscall.Conn should also implement io.Closer.
func tryClose(sc syscall.Conn) error {
	c, ok := sc.(io.Closer)
	if !ok {
		return nil
	}

	return c.Close()
}
