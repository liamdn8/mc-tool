package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/liamdn8/mc-tool/pkg/logger"
	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// TerminalHandler manages websocket connections for the live terminal feature.
type TerminalHandler struct {
	BaseHandler
	service *services.TerminalService
}

// NewTerminalHandler creates a new terminal handler instance.
func NewTerminalHandler(service *services.TerminalService) *TerminalHandler {
	return &TerminalHandler{service: service}
}

type terminalMessage struct {
	Type    string `json:"type"`
	Data    string `json:"data,omitempty"`
	Cols    int    `json:"cols,omitempty"`
	Rows    int    `json:"rows,omitempty"`
	Code    *int   `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

var terminalUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebsocket upgrades the request to a websocket-protected terminal session.
func (h *TerminalHandler) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	conn, err := terminalUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.GetLogger().Error("terminal websocket upgrade failed", map[string]interface{}{"error": err.Error()})
		return
	}
	defer conn.Close()

	conn.SetReadLimit(64 * 1024)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	session, err := h.service.StartSession(ctx, 0, 0)
	if err != nil {
		logger.GetLogger().Error("failed to start terminal session", map[string]interface{}{"error": err.Error()})
		_ = conn.WriteJSON(terminalMessage{Type: "error", Message: "Failed to start shell"})
		return
	}
	defer session.Close()

	var writeMu sync.Mutex
	send := func(msg terminalMessage) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			return err
		}
		return conn.WriteJSON(msg)
	}

	if err := send(terminalMessage{Type: "ready"}); err != nil {
		logger.GetLogger().Error("failed to send terminal ready message", map[string]interface{}{"error": err.Error()})
		return
	}

	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		buf := make([]byte, 4096)
		for {
			n, readErr := session.PTY().Read(buf)
			if n > 0 {
				if err := send(terminalMessage{Type: "output", Data: string(buf[:n])}); err != nil {
					logger.GetLogger().Error("failed to send terminal output", map[string]interface{}{"error": err.Error()})
					return
				}
			}
			if readErr != nil {
				if !errors.Is(readErr, io.EOF) {
					logger.GetLogger().Error("terminal session read error", map[string]interface{}{"error": readErr.Error()})
				}
				return
			}
		}
	}()

	exitDone := make(chan struct{})
	go func() {
		defer close(exitDone)
		waitErr := session.Wait()
		code := 0
		if waitErr != nil {
			if exitErr, ok := waitErr.(*exec.ExitError); ok {
				code = exitErr.ExitCode()
			}
		}
		_ = send(terminalMessage{Type: "exit", Code: &code})
	}()

readLoop:
	for {
		_, payload, readErr := conn.ReadMessage()
		if readErr != nil {
			break
		}

		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg terminalMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			logger.GetLogger().Warn("invalid terminal message", map[string]interface{}{"error": err.Error()})
			continue
		}

		switch msg.Type {
		case "input":
			if _, err := session.PTY().Write([]byte(msg.Data)); err != nil {
				logger.GetLogger().Error("failed to write to terminal", map[string]interface{}{"error": err.Error()})
				break readLoop
			}
		case "resize":
			if err := session.Resize(msg.Cols, msg.Rows); err != nil {
				logger.GetLogger().Warn("terminal resize failed", map[string]interface{}{"error": err.Error()})
			}
		case "ping":
			if err := send(terminalMessage{Type: "pong"}); err != nil {
				logger.GetLogger().Warn("terminal pong failed", map[string]interface{}{"error": err.Error()})
			}
		default:
			logger.GetLogger().Warn("unknown terminal message type", map[string]interface{}{"type": msg.Type})
		}
	}

	cancel()
	_ = session.Close()

	<-outputDone
	<-exitDone
}
