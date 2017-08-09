package main

import (
	"database/sql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"

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
	"template/index.html",
	"template/aws.html"))

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

type M map[string]interface{}

func renderTemplate(w http.ResponseWriter, tmpl string, keys []int, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", M{"keys": keys, "data": data})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	datas := dbQuery("select * from solar_data where deleted=0 order by created desc")
	var keys []int
	for k := range datas {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	renderTemplate(w, "index", keys, datas)
}

func awsHandler(w http.ResponseWriter, r *http.Request) {
	datas := dbAwsQuery("select * from aws_data where deleted=0 order by created desc")
	var keys []int
	for k := range datas {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	renderTemplate(w, "aws", keys, datas)
}

func dataAwsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		datas := dbAwsQuery("select * from solar_data where deleted=0 order by created desc")
		fmt.Println(datas)
		json.NewEncoder(w).Encode(datas)
	case "POST":
		b, err := ioutil.ReadAll(r.Body)
		checkErr(err)
		fmt.Println(string(b))

		var s awsdata
		err = json.Unmarshal(b, &s)
		checkErr(err)
		fmt.Println(s)

		data := dbAwsInsert(s)
		fmt.Println(data)
	}
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getHandler(w)
	case "POST":
		postHandler(w, r)
	case "PUT":

	case "DELETE":
		dbDeleteAll()
	}
}

func getHandler(w http.ResponseWriter) {
	datas := dbQuery("select * from solar_data where deleted=0 order by created desc")
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

	datas := dbQuery("select * from solar_data where deleted=0 order by id DESC")

	var keys []int
	for k := range datas {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	i := 2
	for _, k := range keys {
		fmt.Println("data: ", datas[k])
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", "A", i), datas[k].Created.Format("2006-01-02 15:04:05"))
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

	awsdatas := dbAwsQuery("select * from aws_data where deleted=0 order by id DESC")

	for k := range awsdatas {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	i = 2
	for _, k := range keys {
		fmt.Println("data: ", awsdatas[k])
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "A", i), awsdatas[k].Created.Format("2006-01-02 15:04:05"))
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "B", i), awsdatas[k].IndoorTemp)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "C", i), awsdatas[k].IndoorHumid)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "D", i), awsdatas[k].AbsolutePressure)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "E", i), awsdatas[k].RelativePressure)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "F", i), awsdatas[k].OutdoorTemp)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "G", i), awsdatas[k].OutdoorHumid)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "H", i), awsdatas[k].WindDirection)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "I", i), awsdatas[k].WindSpeed)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "J", i), awsdatas[k].WindGust)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "K", i), awsdatas[k].SolarRadiation)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "L", i), awsdatas[k].UV)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "M", i), awsdatas[k].UVI)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "N", i), awsdatas[k].HourlyRain)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "O", i), awsdatas[k].DailyRain)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "P", i), awsdatas[k].WeeklyRain)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "Q", i), awsdatas[k].MonthlyRain)
		xlsx.SetCellValue("Sheet2", fmt.Sprintf("%s%d", "R", i), awsdatas[k].YearlyRain)
		i++
	}

	err := xlsx.SaveAs("./Solar-Simulator-Exported.xlsx")
	checkErr(err)

	info, _ := os.Stat("./Solar-Simulator-Exported.xlsx")
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
	http.HandleFunc("/data/aws", dataAwsHandler)
	http.HandleFunc("/debug", debugHandler)
	http.HandleFunc("/aws", awsHandler)
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
