package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type qmp struct {
	mu   sync.Mutex
	conn net.Conn
	dec  *json.Decoder
	enc  *json.Encoder
}

type qmpResponse struct {
	Return json.RawMessage `json:"return"`
	Error  *struct {
		Class string `json:"class"`
		Desc  string `json:"desc"`
	} `json:"error"`
	Event string `json:"event"`
}

func dialQMP(path string, timeout time.Duration) (*qmp, error) {
	deadline := time.Now().Add(timeout)
	var conn net.Conn
	for {
		var err error
		if conn, err = net.Dial("unix", path); err == nil {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("connecting to QMP socket %s: %w", path, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	q := &qmp{conn: conn, dec: json.NewDecoder(conn), enc: json.NewEncoder(conn)}
	var greeting map[string]any
	if err := q.dec.Decode(&greeting); err != nil {
		conn.Close()
		return nil, fmt.Errorf("reading QMP greeting: %w", err)
	}
	if _, err := q.execute("qmp_capabilities", nil); err != nil {
		conn.Close()
		return nil, err
	}
	return q, nil
}

func (q *qmp) execute(command string, args any) (json.RawMessage, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	req := map[string]any{"execute": command}
	if args != nil {
		req["arguments"] = args
	}
	if err := q.enc.Encode(req); err != nil {
		return nil, fmt.Errorf("sending QMP %s: %w", command, err)
	}
	for {
		var resp qmpResponse
		if err := q.dec.Decode(&resp); err != nil {
			return nil, fmt.Errorf("reading QMP %s response: %w", command, err)
		}
		switch {
		case resp.Event != "":
			continue // asynchronous event, not our response
		case resp.Error != nil:
			return nil, fmt.Errorf("QMP %s: %s: %s", command, resp.Error.Class, resp.Error.Desc)
		default:
			return resp.Return, nil
		}
	}
}

func (q *qmp) close() error {
	return q.conn.Close()
}

const tabletMax = 0x7fff

type inputEvent map[string]any

func absEvent(axis string, value int) inputEvent {
	return inputEvent{"type": "abs", "data": map[string]any{"axis": axis, "value": value}}
}

func btnEvent(button string, down bool) inputEvent {
	return inputEvent{"type": "btn", "data": map[string]any{"button": button, "down": down}}
}

func keyEvent(qcode string, down bool) inputEvent {
	return inputEvent{"type": "key", "data": map[string]any{
		"down": down,
		"key":  map[string]any{"type": "qcode", "data": qcode},
	}}
}

func (q *qmp) sendInput(events ...inputEvent) error {
	_, err := q.execute("input-send-event", map[string]any{"events": events})
	return err
}

func (q *qmp) screendump(path string) error {
	_, err := q.execute("screendump", map[string]any{"filename": path, "format": "png"})
	return err
}

func (q *qmp) quit() error {
	_, err := q.execute("quit", nil)
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
