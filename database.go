package main

import (
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
	lum2    float,
	deleted bool);
	
	create table solar_debug (id integer primary key auto_increment, 
	created datetime, 
	message text,
	deleted bool);

	create table aws_data(id integer primary key auto_increment,
	created datetime,
	indoor_temp float, indoor_humid float,
	absolute_pressure float, relative_pressure float,
	outdoor_temp float, outdoor_humid float,
	wind_direction float, wind_speed float,
	wind_gust float, solar_radiation float,
	uv float, uvi float,
	hourly_rain_rate float, daily_rain float,
	weekly_rain float, monthly_rain float,
	yearly_rain float,
	deleted bool default 0);
	`
	_, err := db.Exec(sqlStmt)
	checkErr(err)
}

func dbAwsInsert(s awsdata) awsdata {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("insert into aws_data(created, indoor_temp, indoor_humid, absolute_pressure, relative_pressure, outdoor_temp, outdoor_humid, wind_direction, wind_speed, wind_gust, solar_radiation, uv, uvi, hourly_rain_rate, daily_rain, weekly_rain, monthly_rain, yearly_rain) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Stmt(stmt).Exec(time.Now(), s.IndoorTemp, s.IndoorHumid, s.AbsolutePressure, s.RelativePressure, s.OutdoorTemp, s.OutdoorHumid, s.WindDirection, s.WindSpeed, s.WindGust, s.SolarRadiation, s.UV, s.UVI, s.HourlyRain, s.DailyRain, s.WeeklyRain, s.MonthlyRain, s.YearlyRain)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	rows, err := db.Query("select * from aws_data order by created desc limit 1")
	checkErr(err)
	defer rows.Close()

	var data awsdata
	if rows.Next() {
		var id int
		var Date time.Time
		var IndoorTemp float64
		var IndoorHumid float64
		var AbsolutePressure float64
		var RelativePressure float64
		var OutdoorTemp float64
		var OutdoorHumid float64
		var WindDirection float64
		var WindSpeed float64
		var WindGust float64
		var SolarRadiation float64
		var UV float64
		var UVI float64
		var HourlyRain float64
		var DailyRain float64
		var WeeklyRain float64
		var MonthlyRain float64
		var YearlyRain float64
		var Deleted bool
		err = rows.Scan(&id, &Date, &IndoorTemp, &IndoorHumid, &AbsolutePressure, &RelativePressure, &OutdoorTemp, &OutdoorHumid, &WindDirection, &WindSpeed, &WindGust, &SolarRadiation, &UV, &UVI, &HourlyRain, &DailyRain, &WeeklyRain, &MonthlyRain, &YearlyRain, &Deleted)
		if err != nil {
			log.Fatal(err)
		}
		data = awsdata{
			id, Date, IndoorTemp, IndoorHumid, AbsolutePressure, RelativePressure, OutdoorTemp, OutdoorHumid, WindDirection, WindSpeed, WindGust, SolarRadiation, UV, UVI, HourlyRain, DailyRain, WeeklyRain, MonthlyRain, YearlyRain, Deleted,
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
		var Deleted bool
		err = rows.Scan(&id, &Date, &Message, &Deleted)
		if err != nil {
			log.Fatal(err)
		}
		data = solardebug{
			id, Date, Message, Deleted,
		}
	}
	return data
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
		var Deleted bool
		err = rows.Scan(&id, &Date, &Voltage, &Current, &Temperature1, &Temperature2, &LightIntensity1, &LightIntensity2, &Deleted)
		if err != nil {
			log.Fatal(err)
		}
		data = solardata{
			id, Date, Voltage, Current, Temperature1, Temperature2, LightIntensity1, LightIntensity2, Deleted,
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
		var Deleted bool
		err = rows.Scan(&id, &Date, &Voltage, &Current, &Temperature1, &Temperature2, &LightIntensity1, &LightIntensity2, &Deleted)
		if err != nil {
			log.Fatal(err)
		}
		datas[id] = solardata{
			id, Date, Voltage, Current, Temperature1, Temperature2, LightIntensity1, LightIntensity2, Deleted,
		}
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
		var Deleted bool
		err = rows.Scan(&id, &Date, &Message, &Deleted)
		if err != nil {
			log.Fatal(err)
		}
		datas[id] = solardebug{
			id, Date, Message, Deleted,
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return datas
}

func dbAwsQuery(query string) map[int]awsdata {
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	datas := make(map[int]awsdata)
	for rows.Next() {
		var id int
		var Date time.Time
		var IndoorTemp float64
		var IndoorHumid float64
		var AbsolutePressure float64
		var RelativePressure float64
		var OutdoorTemp float64
		var OutdoorHumid float64
		var WindDirection float64
		var WindSpeed float64
		var WindGust float64
		var SolarRadiation float64
		var UV float64
		var UVI float64
		var HourlyRain float64
		var DailyRain float64
		var WeeklyRain float64
		var MonthlyRain float64
		var YearlyRain float64
		var Deleted bool
		err = rows.Scan(&id, &Date, &IndoorTemp, &IndoorHumid, &AbsolutePressure, &RelativePressure, &OutdoorTemp, &OutdoorHumid, &WindDirection, &WindSpeed, &WindGust, &SolarRadiation, &UV, &UVI, &HourlyRain, &DailyRain, &WeeklyRain, &MonthlyRain, &YearlyRain, &Deleted)
		if err != nil {
			log.Fatal(err)
		}
		datas[id] = awsdata{
			id, Date, IndoorTemp, IndoorHumid, AbsolutePressure, RelativePressure, OutdoorTemp, OutdoorHumid, WindDirection, WindSpeed, WindGust, SolarRadiation, UV, UVI, HourlyRain, DailyRain, WeeklyRain, MonthlyRain, YearlyRain, Deleted,
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return datas
}

func dbDeleteAll() {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("update solar_data set deleted=?")
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Stmt(stmt).Exec(true)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}
