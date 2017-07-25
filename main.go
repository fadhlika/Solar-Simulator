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

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/xuri/excelize"
)

var templates = template.Must(template.ParseFiles(
	"template/head.html",
	"template/topbar.html",
	"template/index.html"))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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
	json.NewEncoder(w).Encode(datas)
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
		json.NewEncoder(w).Encode(datas)
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

	}
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
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
	db, err = sql.Open("mysql", "root:saxifrage@/solar_simulator?parseTime=true&loc=Asia%2FJakarta")
	checkErr(err)
	defer db.Close()
	log.Println("Database opened")

	log.Println("Starting websocket message handler")
	go handleMessages()
	go handleDebugMessages()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
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
