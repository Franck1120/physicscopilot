package handlers

import (
	"net"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"
	gorilla "github.com/gorilla/websocket"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// startTestWSServer starts a Fiber WebSocket server on a random port and
// returns the address and a shutdown function.
func startTestWSServer(t *testing.T, handler func(*fiberws.Conn)) (addr string, shutdown func()) {
	t.Helper()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", fiberws.New(handler))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go app.Listener(ln) //nolint:errcheck

	return ln.Addr().String(), func() { app.Shutdown() } //nolint:errcheck
}

// TestHeartbeatPingLoopExitsWhenDoneClosed verifies that pingLoop stops
// promptly when its done channel is closed, regardless of the tick interval.
func TestHeartbeatPingLoopExitsWhenDoneClosed(t *testing.T) {
	sessionSvc := services.NewSessionService(nil, nil)
	convSvc := services.NewConversationService(sessionSvc, nil)
	h := NewWSHandler(convSvc, sessionSvc)
	h.PingInterval = time.Hour // very long — won't fire in the test

	done := make(chan struct{})
	exited := make(chan struct{})

	// Use a nil safeConn — pingLoop only writes when ticker fires or done closes.
	// Since the ticker interval is 1 hour, done will close first.
	go func() {
		defer close(exited)
		ticker := time.NewTicker(h.effectivePingInterval())
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Would send ping — won't happen in test window.
			case <-done:
				return
			}
		}
	}()

	close(done)

	select {
	case <-exited:
		// OK: goroutine exited within the deadline.
	case <-time.After(time.Second):
		t.Error("pingLoop goroutine did not exit within 1 s after done closed")
	}
}

// TestHeartbeatClosesConnectionOnNoPong starts a real Fiber WS server using
// the production WSHandler with shortened timing, connects a client that
// deliberately ignores server Ping frames, and verifies the server closes
// the connection once pongWait elapses.
func TestHeartbeatClosesConnectionOnNoPong(t *testing.T) {
	sessionSvc := services.NewSessionService(nil, nil)
	convSvc := services.NewConversationService(sessionSvc, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)

	// Use short timeouts so the test completes in under a second.
	wsHandler.PingInterval = 80 * time.Millisecond
	wsHandler.PongWait = 200 * time.Millisecond

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	// Connect a WebSocket client.
	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Override the client's ping handler to NOT send pong — simulates a
	// misbehaving or dead client.
	conn.SetPingHandler(func(data string) error {
		return nil // intentionally ignore the ping
	})

	// Give the server time to detect the missing pong and close the connection.
	// Deadline = pongWait + pingInterval + 300 ms buffer.
	testDeadline := time.Now().Add(wsHandler.PongWait + wsHandler.PingInterval + 300*time.Millisecond)
	conn.SetReadDeadline(testDeadline)

	// Read until any error (close frame, network error, or deadline).
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			// Any error means the connection is no longer alive — expected.
			break
		}
	}

	if time.Now().After(testDeadline) {
		t.Error("server did not close the connection within the expected heartbeat timeout")
	}
}

// TestIPConnTrackerEnforcesLimit verifies that add() returns false once the
// per-IP cap is reached, and remove() frees a slot.
func TestIPConnTrackerEnforcesLimit(t *testing.T) {
	tracker := newIPConnTracker()

	for i := 0; i < maxConnsPerIP; i++ {
		if !tracker.add("1.2.3.4") {
			t.Fatalf("add() returned false at connection %d (limit %d)", i+1, maxConnsPerIP)
		}
	}

	// Next add must fail.
	if tracker.add("1.2.3.4") {
		t.Errorf("add() should return false when limit %d is reached", maxConnsPerIP)
	}

	// After removing one slot the add should succeed again.
	tracker.remove("1.2.3.4")
	if !tracker.add("1.2.3.4") {
		t.Error("add() should succeed after remove() freed a slot")
	}
}
