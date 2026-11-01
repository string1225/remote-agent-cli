package gitauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func startFixtureServer(t *testing.T, server *Server) net.Conn {
	t.Helper()
	client, service := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); server.Serve(ctx, service) }()
	t.Cleanup(func() {
		cancel()
		client.Close()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("connection handler leaked")
		}
	})
	return client
}

func TestProtocolStreamsOnlyToAuthorizedClients(t *testing.T) {
	server := &Server{Manager: fixtureManager(t), Authorize: func(net.Conn) error { return nil }}
	conn := startFixtureServer(t, server)
	var output bytes.Buffer
	if err := Exchange(context.Background(), conn, Request{Version: Version, Action: "login"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "TEST-CODE") {
		t.Fatal("device code was not streamed")
	}
	server = &Server{Manager: fixtureManager(t), Authorize: func(net.Conn) error { return errors.New("caller denied") }}
	conn = startFixtureServer(t, server)
	if err := Exchange(context.Background(), conn, Request{Version: Version, Action: "status"}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatal("unauthorized client was not rejected", err)
	}
}

func TestProtocolRejectsUnknownFieldsAndVersions(t *testing.T) {
	for _, input := range []string{`{"version":1,"action":"login","password":"must-not-echo"}`, `{"version":200,"action":"login"}`} {
		server := &Server{Manager: fixtureManager(t), Authorize: func(net.Conn) error { return nil }}
		conn := startFixtureServer(t, server)
		go io.WriteString(conn, input+"\n")
		var event Event
		if err := json.NewDecoder(conn).Decode(&event); err != nil || !event.Done || event.Error == "" {
			t.Fatal(event, err)
		}
		if strings.Contains(event.Error, "must-not-echo") {
			t.Fatal("request content leaked")
		}
	}
}

func TestProtocolCancellationTerminatesHelper(t *testing.T) {
	server := &Server{Manager: fixtureManager(t), Authorize: func(net.Conn) error { return nil }}
	conn := startFixtureServer(t, server)
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if err := Exchange(ctx, conn, Request{Version: Version, Action: "login", Username: "wait"}, &bytes.Buffer{}); err == nil {
		t.Fatal("cancelled request succeeded")
	}
}

func TestProtocolBoundsOutputAndSerializesOperations(t *testing.T) {
	var wire bytes.Buffer
	w := &eventWriter{encoder: json.NewEncoder(&wire)}
	if _, err := w.Write(bytes.Repeat([]byte("x"), MaxOutput+1)); err == nil {
		t.Fatal("unbounded helper output accepted")
	}
	if wire.Len() > MaxOutput+4096 {
		t.Fatal("output exceeded wire bound")
	}
	server := &Server{Manager: fixtureManager(t), Authorize: func(net.Conn) error { return nil }}
	server.busy.Lock()
	defer server.busy.Unlock()
	conn := startFixtureServer(t, server)
	if err := Exchange(context.Background(), conn, Request{Version: Version, Action: "status"}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "in progress") {
		t.Fatal("concurrent operation not rejected", err)
	}
}
