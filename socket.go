package socket

import (
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Socket struct {
	c           *websocket.Conn
	mu          sync.Mutex
	pongMessage string
}

type JSON map[string]interface{}

func New() *Socket {
	return &Socket{}
}

/*
	Connect connects to the websocket server.
	scheme e.g. wss
	host e.g. leagueoftraders.io
	pingInterval how long to send the ping message
	ping message used to ping the server e.g. {"op":"ping"}
	pong message used to pong the server e.g. {"type": "pong"}
*/
func (s *Socket) Connect(scheme, host, path string, pingInterval time.Duration, pingMessage, pongMessage string) error {
	u := url.URL{Scheme: scheme, Host: host, Path: path}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	s.c = conn
	s.pongMessage = pongMessage
	if pingInterval > 0 {
		go func() {
			for range time.NewTicker(pingInterval).C {
				if s.c == nil {
					break
				}
				err = s.Send(pingMessage)
				if err != nil {
					break
				}
			}
		}()
	}
	return nil
}

func (s *Socket) Send(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.c.WriteMessage(websocket.TextMessage, []byte(message))
}

func (s *Socket) SendJson(j JSON) error {
	data, _ := json.Marshal(j)
	return s.Send(string(data))
}

func (s *Socket) Read() ([]byte, error) {
	_, msg, err := s.c.ReadMessage()
	if err != nil {
		return nil, err
	}
	if len(msg) == len(s.pongMessage) && string(msg) == s.pongMessage {
		return s.Read()
	}
	return msg, nil
}

func (s *Socket) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.Close()
}
