package gitauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

const MaxOutput = 64 << 10

type Event struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Done    bool   `json:"done,omitempty"`
}

type Server struct {
	Manager   Manager
	Manage    func(context.Context, json.RawMessage) (any, error)
	Authorize func(net.Conn) error
	busy      sync.Mutex
}

type eventWriter struct {
	mu      sync.Mutex
	encoder *json.Encoder
	total   int
}

func (w *eventWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.total+len(data) > MaxOutput {
		return 0, errors.New("Git authorization output limit exceeded")
	}
	w.total += len(data)
	n := len(data)
	for len(data) > 0 {
		size := min(len(data), 4096)
		if err := w.encoder.Encode(Event{Message: string(data[:size])}); err != nil {
			return 0, err
		}
		data = data[size:]
	}
	return n, nil
}

// Serve never logs command output or request bodies. Only this local connection
// receives the short-lived device code; disconnecting it cancels the helper.
func (s *Server) Serve(parent context.Context, conn net.Conn) {
	ctx, cancel := context.WithTimeout(parent, 15*time.Minute)
	defer cancel()
	defer conn.Close()
	go func(done <-chan struct{}) { <-done; conn.Close() }(ctx.Done())
	encoder := json.NewEncoder(conn)
	finish := func(err error) {
		event := Event{Done: true}
		if err != nil {
			event.Error = err.Error()
		}
		_ = encoder.Encode(event)
	}
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	d := json.NewDecoder(io.LimitReader(conn, 32<<10))
	d.DisallowUnknownFields()
	var request Request
	if err := d.Decode(&request); err != nil || request.Version != Version {
		finish(errors.New("invalid Git authorization request"))
		return
	}
	// A successful pipe read establishes the identification token used by the
	// Windows authorizer. No helper or account output occurs before authorization.
	if s.Authorize == nil {
		finish(errors.New("Git authorization caller is not authorized"))
		return
	}
	if err := s.Authorize(conn); err != nil {
		finish(err)
		return
	}
	if err := request.Validate(); err != nil {
		finish(err)
		return
	}
	if request.Action == "manage" {
		if s.Manage == nil {
			finish(errors.New("请升级客户端以启用服务器和 GitHub 管理"))
			return
		}
		_ = conn.SetDeadline(time.Now().Add(90 * time.Second))
		result, err := s.Manage(ctx, request.Body)
		if err == nil {
			raw, e := json.Marshal(result)
			err = e
			if err == nil {
				_, err = (&eventWriter{encoder: encoder}).Write(raw)
			}
		}
		finish(err)
		return
	}
	if !s.busy.TryLock() {
		finish(errors.New("another Git authorization operation is in progress"))
		return
	}
	defer s.busy.Unlock()
	if request.Action != "login" {
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, 30*time.Second)
		defer stop()
	}
	deadline, _ := ctx.Deadline()
	_ = conn.SetDeadline(deadline)
	go func() { var extra [1]byte; _, _ = conn.Read(extra[:]); cancel() }()
	w := &eventWriter{encoder: encoder}
	err := s.Manager.Run(ctx, request, w)
	if err == nil {
		switch request.Action {
		case "status":
			if w.total == 0 {
				_, err = w.Write([]byte("No GitHub account is authorized for the Windows service. Run remote-agent git login.\n"))
			}
		case "login":
			_, err = w.Write([]byte("GitHub authorization saved for Windows service tasks.\n"))
		case "logout":
			_, err = w.Write([]byte("Selected GitHub account removed from the Windows service credential store.\n"))
		}
	}
	finish(err)
}

func Exchange(ctx context.Context, conn net.Conn, request Request, output io.Writer) error {
	defer conn.Close()
	if err := request.Validate(); err != nil {
		return err
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()
	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return err
	}
	d := json.NewDecoder(io.LimitReader(conn, MaxOutput*8))
	for {
		var event Event
		if err := d.Decode(&event); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return errors.New("Git authorization connection ended; retry remote-agent git status")
		}
		if event.Message != "" {
			if _, err := io.WriteString(output, event.Message); err != nil {
				return err
			}
		}
		if event.Done {
			if event.Error != "" {
				return errors.New(event.Error)
			}
			return nil
		}
	}
}
