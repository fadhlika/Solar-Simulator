package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pongPeriod = (pongWait * 9) / 10
)

var url string = "127.0.0.1:8000"

type Clients struct {
	c  map[*websocket.Conn]bool
	mu sync.Mutex
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panic(err)
	}
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	clients.c[ws] = true
	for {
		var data solardata
		err := ws.ReadJSON(&data)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error close: %v", err)
			}
			log.Printf("error: %v", err)
			delete(clients.c, ws)
			break
		}
		broadcast <- data
	}
}

func handleMessages() {
	clients.mu.Lock()
	defer clients.mu.Unlock()

	ticker := time.NewTicker(pongPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-broadcast:
			for client := range clients.c {
				defer client.Close()
				client.SetWriteDeadline(time.Now().Add(writeWait))

				if !ok {
					client.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := client.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Printf("error nextwriter: %v", err)
					client.Close()
					delete(clients.c, client)
				}
				err = json.NewEncoder(w).Encode(message)
				if err != nil {
					log.Printf("error json newencoder: %v", err)
					w.Close()
					delete(clients.c, client)
				}

				if err = w.Close(); err != nil {
					return
				}
			}
		case <-ticker.C:
			for client := range clients.c {
				client.SetWriteDeadline(time.Now().Add(writeWait))
				if err := client.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Println("Websocket ping error")
					return
				}
			}
		}
	}
}

func handleDebugConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	debugclients.c[ws] = true
	for {
		_, m, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error close: %v", err)
			}
			log.Printf("error: %v", err)
			delete(debugclients.c, ws)
			break
		}
		debugchannel <- m
	}
}

func handleDebugMessages() {
	debugclients.mu.Lock()
	defer debugclients.mu.Unlock()

	ticker := time.NewTicker(pongPeriod)
	defer ticker.Stop()

	for {
		select {
		case m, ok := <-debugchannel:
			for client := range debugclients.c {
				defer client.Close()
				client.SetWriteDeadline(time.Now().Add(writeWait))

				if !ok {
					client.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := client.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Printf("error nextwriter: %v", err)
					client.Close()
					delete(clients.c, client)
				}
				_, err = w.Write([]byte(m))
				if err != nil {
					log.Printf("error: %v", err)
					w.Close()
					delete(debugclients.c, client)
				}
				if err = w.Close(); err != nil {
					return
				}
			}
		case <-ticker.C:
			for client := range debugclients.c {
				client.SetWriteDeadline(time.Now().Add(writeWait))
				if err := client.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Println("Websocket ping error")
					return
				}
			}
		}
	}
}

func dialWs() *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+url+"/ws", nil)
	if err != nil {
		log.Println("write: ", err)
	}
	return conn
}

func dialDebugWs() *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+url+"/wsd", nil)
	if err != nil {
		log.Println("write: ", err)
	}
	return conn
}

func sendWS(s solardata) {
	conn := dialWs()
	err := conn.WriteJSON(s)
	if err != nil {
		log.Println("write: ", err)
	}
}

func sendDebugWS(s solardebug) {
	conn := dialDebugWs()
	err := conn.WriteJSON(s)
	if err != nil {
		log.Println("write: ", err)
	}
}
