func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panic(err)
	}
	defer ws.Close()

	clients[ws] = true
	for {
		var data solardata
		err := ws.ReadJSON(&data)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- data
	}
}

func handleMessages() {
	for {
		data := <-broadcast
		for client := range clients {
			err := client.WriteJSON(data)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
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
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", nil)
	if err != nil {
		log.Println("write: ", err)
	}
	return conn
}

func dialDebugWs() *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/wsd", nil)
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