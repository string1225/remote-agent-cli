package webproxy

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	ChunkSize       = 32 * 1024
	MaxPerDevice    = 32
	MaxConnections  = 256
	MaxLifetime     = 30 * time.Minute
	HeartbeatPeriod = 20 * time.Second
)

// Tunnel transports bytes without terminating the target's TLS. reader may be
// a hijacked HTTP reader containing bytes pipelined after CONNECT headers.
// Both directions and the heartbeat stop together on any error or cancellation.
func Tunnel(ctx context.Context, ws *websocket.Conn, conn net.Conn, reader io.Reader, authorized func() bool) {
	ctx, cancel := context.WithCancel(ctx)
	closeBoth := func() { ws.Close(); conn.Close() }
	stop := context.AfterFunc(ctx, closeBoth)
	defer func() { cancel(); closeBoth(); stop() }()
	ws.SetReadLimit(ChunkSize)
	_ = ws.SetReadDeadline(time.Now().Add(65 * time.Second))
	ws.SetPongHandler(func(string) error { return ws.SetReadDeadline(time.Now().Add(65 * time.Second)) })
	var workers sync.WaitGroup
	workers.Go(func() {
		defer cancel()
		buf := make([]byte, ChunkSize)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				_ = ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if ws.WriteMessage(websocket.BinaryMessage, buf[:n]) != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	})
	workers.Go(func() {
		defer cancel()
		ticker := time.NewTicker(HeartbeatPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if authorized != nil && !authorized() {
					return
				}
				if ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)) != nil {
					return
				}
			}
		}
	})
	for {
		kind, message, err := ws.NextReader()
		if err != nil || kind != websocket.BinaryMessage {
			break
		}
		_ = conn.SetWriteDeadline(time.Now().Add(60 * time.Second))
		if _, err := io.Copy(conn, message); err != nil {
			break
		}
	}
	cancel()
	closeBoth()
	workers.Wait()
}

// Pool bounds even pending dials, tracks upgraded connections for shutdown,
// and imposes a finite lease. Its zero value is ready to use.
type Pool struct {
	mu     sync.Mutex
	closed bool
	next   uint64
	calls  map[uint64]poolCall
}

type poolCall struct {
	device string
	cancel context.CancelFunc
}

func (p *Pool) Open(parent context.Context, device string) (context.Context, func(), bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	count := 0
	for _, call := range p.calls {
		if call.device == device {
			count++
		}
	}
	if p.closed || len(p.calls) >= MaxConnections || count >= MaxPerDevice {
		return nil, nil, false
	}
	if p.calls == nil {
		p.calls = make(map[uint64]poolCall)
	}
	p.next++
	id := p.next
	ctx, cancel := context.WithTimeout(parent, MaxLifetime)
	p.calls[id] = poolCall{device: device, cancel: cancel}
	return ctx, func() {
		cancel()
		p.mu.Lock()
		delete(p.calls, id)
		p.mu.Unlock()
	}, true
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	for _, call := range p.calls {
		call.cancel()
	}
}
