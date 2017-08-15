package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pongPeriod = (pongWait * 9) / 10
)

var url string = "ws://128.199.162.40"

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
	clients[ws] = true
	for {
		var data solardata
		err := ws.ReadJSON(&data)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error close: %v", err)
			}
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- data
	}
}

func handleMessages() {
	ticker := time.NewTicker(pongPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-broadcast:
			log.Printf("Websocket message: %v\n", message)
			for client := range clients {
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
					delete(clients, client)
				}
				err = json.NewEncoder(w).Encode(message)
				if err != nil {
					log.Printf("error json newencoder: %v", err)
					w.Close()
					delete(clients, client)
				}

				if err = w.Close(); err != nil {
					return
				}
			}
		case <-ticker.C:
			for client := range clients {
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

	debugclients[ws] = true
	for {
		_, m, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(debugclients, ws)
			break
		}
		debugchannel <- m
	}
}

func handleDebugMessages() {
	for {
		m := <-debugchannel
		for client := range debugclients {
			err := client.WriteMessage(websocket.TextMessage, m)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(debugclients, client)
			}
		}
	}
}

func dialWs() *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial(url+"/ws", nil)
	if err != nil {
		log.Println("write: ", err)
	}
	return conn
}

func dialDebugWs() *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial(url+"/wsd", nil)
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
