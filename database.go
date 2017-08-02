package main

import (
	"fmt"
	"log"
	"time"
)

func dbInit() {
	sqlStmt := `
	create table solar_data (id integer primary key auto_increment, 
		created datetime, 
		voltage float,
		current float,
		temp1   float,
		temp2   float,
		lum1    float,
		lum2    float);
	create table solar_debug (id integer primary key auto_increment, 
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

	_, err = tx.Stmt(stmt).Exec(time.Now(), s.Voltage, s.Current, s.Temp1, s.Temp2, s.Lum1, s.Lum2)
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
