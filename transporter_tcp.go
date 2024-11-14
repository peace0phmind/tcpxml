package tcpxml

import (
	"bytes"
	"context"
	"errors"
	"github.com/expgo/factory"
	"github.com/expgo/serve"
	"github.com/thejerf/suture/v4"
	"net"
	"time"
)

type TcpTransporter struct {
	baseTransporter
	conn    net.Conn
	buf     bytes.Buffer
	lines   chan string
	serveId suture.ServiceToken
}

func NewTcpTransport(addr string) *TcpTransporter {
	ret := factory.NewBeforeInit[TcpTransporter](func(ret *TcpTransporter) {
		ret.baseTransporter.addr = addr
	})

	ret.serveId = serve.AddServe("tcpxml-tcp", ret)
	ret.lines = make(chan string, 10000)
	_ = ret.open()

	return ret
}

func (t *TcpTransporter) open() (err error) {
	if !t.running.CompareAndSwap(false, true) {
		return nil
	}

	if t.state == StateConnected {
		return nil
	}

	t.setState(StateConnecting, nil)
	dialer := net.Dialer{Timeout: 3 * time.Second}
	t.conn, err = dialer.Dial("tcp", t.addr)
	if err != nil {
		t.L.Warnf("DialTCP %s failed: %v", t.addr, err)
		t.setState(StateDisconnected, err)
		return err
	}

	t.setState(StateConnected, nil)

	return err
}

func (t *TcpTransporter) Close() (err error) {
	defer func() {
		t.setState(StateConnectClosed, err)
		t.conn = nil
	}()

	_ = t.baseTransporter.Close()

	if t.conn == nil {
		return nil
	}

	_ = serve.RemoveAndWait(t.serveId, 300*time.Millisecond)

	return t.conn.Close()
}

func (t *TcpTransporter) Write(data []byte) (n int, err error) {
	if t.conn == nil || t.state == StateDisconnected {
		return 0, errors.New("tcp transporter not connected")
	}

	defer func() {
		if err != nil {
			t.setState(StateDisconnected, err)
		}
	}()

	err = t.conn.SetWriteDeadline(time.Now().Add(t.WriteTimeout))
	if err != nil {
		return 0, err
	}

	return t.conn.Write(data)
}

func (t *TcpTransporter) Read(buf []byte) (n int, err error) {
	if t.conn == nil || t.state == StateDisconnected {
		return 0, errors.New("tcp transporter not connected")
	}

	defer func() {
		if err != nil {
			t.setState(StateDisconnected, err)
		}
	}()

	err = t.conn.SetReadDeadline(time.Now().Add(t.ReadTimeout))
	if err != nil {
		return 0, err
	}

	return t.conn.Read(buf)
}

func (t *TcpTransporter) Serve(ctx context.Context) error {
	if err := t.open(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if t.state != StateConnected {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if err := t.conn.SetReadDeadline(time.Now().Add(t.ReadTimeout)); err != nil {
				return err
			}
			var buf [4096]byte
			l, err := t.conn.Read(buf[:])
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// fmt.Println("Read timeout")
				} else {
					return err
				}
			}

			_, err = t.buf.Write(buf[:l])
			if err != nil {
				return err
			}

			for {
				line, err1 := t.buf.ReadString('\n')
				if err1 != nil {
					t.L.Warnf("%+v", err1)
					break
				}
				t.lines <- line
			}
		}
	}
}

func (t *TcpTransporter) Lines() <-chan string {
	return t.lines
}
