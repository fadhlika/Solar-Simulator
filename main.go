package main

import (
	"database/sql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"fmt"

	"encoding/json"

	"os"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize"
)

var templates = template.Must(template.ParseFiles("template/head.html", "template/topbar.html", "template/index.html"))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]bool)
var debugclients = make(map[*websocket.Conn]bool)
var broadcast = make(chan solardata)
var debugchannel = make(chan []byte)

var db *sql.DB

type solardata struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created"`
	Voltage float64   `json:"voltage"`
	Current float64   `json:"current"`
	Temp1   float64   `json:"temp1"`
	Temp2   float64   `json:"temp2"`
	Lum1    float64   `json:"lum1"`
	Lum2    float64   `json:"lum2"`
}

type solardebug struct {
	ID      int
	Created time.Time
	Message string
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index")
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getHandler(w)
	case "POST":
		postHandler(w, r)
	case "PUT":

	case "DELETE":
	}
}

func getHandler(w http.ResponseWriter) {
	datas := dbQuery("select * from solar_data order by created desc")
	fmt.Println(datas)
	w.Header().Set("Content-Type", "application/json")
	p, _ := json.Marshal(datas)
	code, err := w.Write(p)
	fmt.Printf("Code: %i", code)
	checkErr(err)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	//Get json data from POST request body and decode to solardata struct
	log.Println(r.Body)
	b, err := ioutil.ReadAll(r.Body)
	checkErr(err)
	fmt.Println(string(b))

	var s solardata
	err = json.Unmarshal(b, &s)
	checkErr(err)
	fmt.Println(s)

	data := dbInsert(s)
	sendWS(data)
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		datas := dbDebugQuery("select * from solar_debug order by created desc")
		fmt.Println(datas)
		w.Header().Set("Content-Type", "application/json")
		p, _ := json.Marshal(datas)
		w.Write(p)
	case "POST":
		//Get json data from POST request body and decode to solardata struct
		log.Println(r.Body)
		b, err := ioutil.ReadAll(r.Body)
		checkErr(err)
		fmt.Println(string(b))

		data := dbDebugInsert(string(b))
		sendDebugWS(data)
	case "PUT":

	case "DELETE":
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

func dbInit() {
	sqlStmt := `
	create table solar_data (id integer primary key, 
		created datetime, 
		voltage float,
		current float,
		temp1   float,
		temp2   float,
		lum1    float,
		lum2    float);
	create table solar_debug (id integer primary key, 
		created datetime, 
		message text);
	`
	_, err := db.Exec(sqlStmt)
	checkErr(err)
}

func dbInsert(s solardata) solardata {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("insert into solar_data(created, voltage, current, temp1, temp2, lum1, lum2) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Stmt(stmt).Exec(s.Created, s.Voltage, s.Current, s.Temp1, s.Temp2, s.Lum1, s.Lum2)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	rows, err := db.Query("select * from solar_data order by created desc limit 1")
	checkErr(err)
	defer rows.Close()

	var data solardata
	if rows.Next() {
		var id int
		var Date time.Time
		var Voltage float64
		var Current float64
		var Temperature1 float64
		var Temperature2 float64
		var LightIntensity1 float64
		var LightIntensity2 float64
		err = rows.Scan(&id, &Date, &Voltage, &Current, &Temperature1, &Temperature2, &LightIntensity1, &LightIntensity2)
		if err != nil {
			log.Fatal(err)
		}
		data = solardata{
			id, Date, Voltage, Current, Temperature1, Temperature2, LightIntensity1, LightIntensity2,
		}
	}
	return data
}

func dbDebugInsert(s string) solardebug {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("insert into solar_debug(created, message) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Stmt(stmt).Exec(time.Now(), s)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	rows, err := db.Query("select * from solar_debug order by created desc limit 1")
	checkErr(err)
	defer rows.Close()

	var data solardebug
	if rows.Next() {
		var id int
		var Date time.Time
		var Message string
		err = rows.Scan(&id, &Date, &Message)
		if err != nil {
			log.Fatal(err)
		}
		data = solardebug{
			id, Date, Message,
		}
	}
	return data
}

func dbQuery(query string) map[int]solardata {

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	datas := make(map[int]solardata)
	for rows.Next() {
		var id int
		var Date time.Time
		var Voltage float64
		var Current float64
		var Temperature1 float64
		var Temperature2 float64
		var LightIntensity1 float64
		var LightIntensity2 float64
		err = rows.Scan(&id, &Date, &Voltage, &Current, &Temperature1, &Temperature2, &LightIntensity1, &LightIntensity2)
		if err != nil {
			log.Fatal(err)
		}
		datas[id] = solardata{
			id, Date, Voltage, Current, Temperature1, Temperature2, LightIntensity1, LightIntensity2,
		}
		fmt.Println(datas[id])
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return datas
}

func dbDebugQuery(query string) map[int]solardebug {
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	datas := make(map[int]solardebug)
	for rows.Next() {
		var id int
		var Date time.Time
		var Message string
		err = rows.Scan(&id, &Date, &Message)
		if err != nil {
			log.Fatal(err)
		}
		datas[id] = solardebug{
			id, Date, Message,
		}
		fmt.Println(datas[id])
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return datas
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
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

func exportHandler(w http.ResponseWriter, r *http.Request) {
	os.Remove("./Solar-Simulator-Exported.xlsx")
	xlsx := excelize.NewFile()
	xlsx.SetCellValue("Sheet1", "A1", "Date")
	xlsx.SetCellValue("Sheet1", "B1", "Voltage")
	xlsx.SetCellValue("Sheet1", "C1", "Current")
	xlsx.SetCellValue("Sheet1", "D1", "Temp1")
	xlsx.SetCellValue("Sheet1", "E1", "Temp2")
	xlsx.SetCellValue("Sheet1", "F1", "Lum1")
	xlsx.SetCellValue("Sheet1", "G1", "Lum2")

	datas := dbQuery("select * from solar_data order by id DESC")

	var keys []int
	for k := range datas {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	fmt.Println(datas)
	i := 2
	for _, k := range keys {
		fmt.Println("data: ", datas[k])
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", "A", i), datas[k].Created)
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", "B", i), datas[k].Voltage)
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", "C", i), datas[k].Current)
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", "D", i), datas[k].Temp1)
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", "E", i), datas[k].Temp2)
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", "F", i), datas[k].Lum1)
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", "G", i), datas[k].Lum2)
		i++
	}

	i--
	xlsx.AddChart("Sheet1", "I2", fmt.Sprintf(`{"type": "scatter", "series":[
		{"name":"=Sheet1!$B$1","categories":"=Sheet1!$A$2:$A$%d","values":"=Sheet1!$B$2:$B$%d"},
		{"name":"=Sheet1!$C$1","categories":"=Sheet1!$A$2:$A$%d","values":"=Sheet1!$C$2:$C$%d"}
		], "title":{"name": "Voltage"}}`, i, i, i, i))

	xlsx.AddChart("Sheet1", "I17", fmt.Sprintf(`{"type": "scatter", "series":[
		{"name":"=Sheet1!$D$1","categories":"=Sheet1!$A$2:$A$%d","values":"=Sheet1!$D$2:$D$%d"},
		{"name":"=Sheet1!$E$1","categories":"=Sheet1!$A$2:$A$%d","values":"=Sheet1!$E$2:$E$%d"}
		], "title":{"name": "Temperature"}}`, i, i, i, i))

	xlsx.AddChart("Sheet1", "R2", fmt.Sprintf(`{"type": "scatter", "series":[
		{"name":"=Sheet1!$F$1","categories":"=Sheet1!$A$2:$A$%d","values":"=Sheet1!$F$2:$F$%d"},
		{"name":"=Sheet1!$G$1","categories":"=Sheet1!$A$2:$A$%d","values":"=Sheet1!$G$2:$G$%d"}
		], "title":{"name": "Luminance"}}`, i, i, i, i))

	err := xlsx.SaveAs("./Solar-Simulator-Exported.xlsx")
	info, _ := os.Stat("./Solar-Simulator-Exported.xlsx")
	checkErr(err)
	fmt.Printf("excel saved, size: %d bytes\r\n", info.Size())

	http.ServeFile(w, r, "./Solar-Simulator-Exported.xlsx")
}

func main() {
	log.Println("starting application")
	log.Println("Opening database")
	var err error
	db, err = sql.Open("sqlite3", "./solar-simulator.db?loc=auto")
	checkErr(err)
	defer db.Close()
	log.Println("Database opened")

	if _, err := os.Stat("./solar-simulator.db"); os.IsNotExist(err) {
		dbInit()
		log.Println("Initializing Database")
	}

	log.Println("Starting websocket message handler")
	go handleMessages()
	go handleDebugMessages()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/data", dataHandler)
	http.HandleFunc("/debug", debugHandler)
	http.HandleFunc("/export", exportHandler)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/wsd", handleDebugConnections)
	/*
		conn, err := net.Dial("tcp", "8.8.8.8:80")
		checkErr(err)
		defer conn.Close()

		localAddr := conn.LocalAddr().String()
		log.Printf("Application started in http://%s:8000", strings.Split(localAddr, ":")[0])
	*/

	log.Println("Application started in http://127.0.0.1:8000")
	http.ListenAndServe(":8000", nil)
}
