package main

import "time"
import "log"

// Solardata struct for data from scemo
type Solardata struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created"`
	Voltage float64   `json:"voltage"`
	Current float64   `json:"current"`
	Temp1   float64   `json:"temp1"`
	Temp2   float64   `json:"temp2"`
	Lum1    float64   `json:"lum1"`
	Lum2    float64   `json:"lum2"`
	PWM		float64	  `json:"pwm"`
	Deleted bool      `json:"deleted"`
}

// Solardebug debug message from esp for debug purpose
type Solardebug struct {
	ID      int
	Created time.Time
	Message string
	Deleted bool
}

// Awsdata struct for data from aws
type Awsdata struct {
	ID               int       `json:"id"`
	Created          time.Time `json:"created"`
	IndoorTemp       float64   `json:"indoortemp"`
	IndoorHumid      float64   `json:"indoorhumid"`
	AbsolutePressure float64   `json:"absolutepressure"`
	RelativePressure float64   `json:"relativepressure"`
	OutdoorTemp      float64   `json:"outdoortemp"`
	OutdoorHumid     float64   `json:"outdoorhumid"`
	WindDirection    float64   `json:"winddirection"`
	WindSpeed        float64   `json:"windspeed"`
	WindGust         float64   `json:"windgust"`
	SolarRadiation   float64   `json:"solarradiation"`
	UV               float64   `json:"uv"`
	UVI              float64   `json:"uvi"`
	HourlyRain       float64   `json:"hourlyrain"`
	DailyRain        float64   `json:"dailyrain"`
	WeeklyRain       float64   `json:"weeklyrain"`
	MonthlyRain      float64   `json:"monthlyrain"`
	YearlyRain       float64   `json:"yearlyrain"`
	Deleted          bool      `json:"deleted"`
}

// CreateAll create all table
func CreateAll() error {
	sqlStmt := `
	create table solar_data (id integer primary key auto_increment, 
	created datetime, 
	voltage float,
	current float,
	temp1   float,
	temp2   float,
	lum1    float,
	lum2    float,
	deleted bool default 0);
	
	create table solar_debug (id integer primary key auto_increment, 
	created datetime, 
	message text,
	deleted bool default 0);

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
	if err != nil {
		log.Printf("Error execute statement %v \n", err.Error())
		return err
	}
	return nil
}

// QuerySolarData get all data in the database and error
func QuerySolarData() (map[int]Solardata, error) {
	rows, err := db.Query("select * from solar_data where deleted=0 order by created desc")
	if err != nil {
		log.Printf("Query data err %v \n", err.Error())
		return nil, err
	}
	defer rows.Close()

	datas := make(map[int]Solardata)
	for rows.Next() {
		var id int
		var Date time.Time
		var Voltage float64
		var Current float64
		var Temperature1 float64
		var Temperature2 float64
		var LightIntensity1 float64
		var LightIntensity2 float64
		var PWM float64
		var Deleted bool
		err = rows.Scan(&id, &Date, &Voltage, &Current, &Temperature1, &Temperature2, &LightIntensity1, &LightIntensity2, &PWM, &Deleted)
		if err != nil {
			log.Printf("Query scan err %v \n", err.Error())
			return nil, err
		}
		datas[id] = Solardata{
			id, Date, Voltage, Current, Temperature1, Temperature2, LightIntensity1, LightIntensity2, PWM, Deleted,
		}
	}
	err = rows.Err()
	if err != nil {
		log.Printf("Query data err %v \n", err.Error())
		return nil, err
	}
	return datas, nil
}

func (d Solardata) save() error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error begin database %v\n", err.Error())
		return err
	}

	stmt, err := db.Prepare("insert into solar_data(created, voltage, current, temp1, temp2, lum1, lum2, pwm) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error prepare statement %v\n", err.Error())
		return err
	}

	res, err := tx.Stmt(stmt).Exec(d.Created, d.Voltage, d.Current, d.Temp1, d.Temp2, d.Lum1, d.Lum2, d.PWM)
	if err != nil {
		log.Printf("Error execute statement %v\n", err.Error())
		return err
	}
	if err = tx.Commit(); err != nil {
		log.Printf("Error commit database %v\n", err.Error())
		return err
	}

	row, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error result %v\n", err.Error())
		return err
	}
	log.Printf("Row affected %v\n", row)

	return nil
}

// QueryDebug returns solardebug and error
func QueryDebug() (map[int]Solardebug, error) {
	rows, err := db.Query("select * from solar_debug order by created desc")
	if err != nil {
		log.Printf("Query debug err %v \n", err.Error())
		return nil, err
	}
	defer rows.Close()

	datas := make(map[int]Solardebug)
	for rows.Next() {
		var id int
		var Date time.Time
		var Message string
		var Deleted bool
		err = rows.Scan(&id, &Date, &Message, &Deleted)
		if err != nil {
			log.Printf("Query scan err %v \n", err.Error())
			return nil, err
		}
		datas[id] = Solardebug{
			id, Date, Message, Deleted,
		}
	}
	err = rows.Err()
	if err != nil {
		log.Printf("Query error %v\n", err.Error())
		return nil, err
	}
	return datas, err
}

func (d Solardebug) save() error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error begin database %v\n", err.Error())
		return err
	}

	stmt, err := db.Prepare("insert into solar_debug(created, message) values(?, ?)")
	if err != nil {
		log.Printf("Error prepare database %v\n", err.Error())
		return err
	}

	_, err = tx.Stmt(stmt).Exec(d.Created, d.Message)
	if err != nil {
		log.Printf("Error execute statement %v\n", err.Error())
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error commit database %v\n", err.Error())
		return err
	}

	return err
}

// QueryAwsData return aws data in database and error
func QueryAwsData() (map[int]Awsdata, error) {
	rows, err := db.Query("select * from aws_data where deleted=0 order by created desc")
	if err != nil {
		log.Printf("Query aws data err %v \n", err.Error())
		return nil, err
	}
	defer rows.Close()

	datas := make(map[int]Awsdata)
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
			log.Printf("Query scan err %v \n", err.Error())
			return nil, err
		}
		datas[id] = Awsdata{
			id, Date, IndoorTemp, IndoorHumid, AbsolutePressure, RelativePressure, OutdoorTemp, OutdoorHumid, WindDirection, WindSpeed, WindGust, SolarRadiation, UV, UVI, HourlyRain, DailyRain, WeeklyRain, MonthlyRain, YearlyRain, Deleted,
		}
	}
	err = rows.Err()
	if err != nil {
		log.Printf("Query aws data err %v \n", err.Error())
		return nil, err
	}
	return datas, err
}

func (s Awsdata) save() error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error begin database %v\n", err.Error())
		return err
	}

	stmt, err := db.Prepare("insert into aws_data(created, indoor_temp, indoor_humid, absolute_pressure, relative_pressure, outdoor_temp, outdoor_humid, wind_direction, wind_speed, wind_gust, solar_radiation, uv, uvi, hourly_rain_rate, daily_rain, weekly_rain, monthly_rain, yearly_rain) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error begin database %v\n", err.Error())
		return err
	}

	_, err = tx.Stmt(stmt).Exec(time.Now(), s.IndoorTemp, s.IndoorHumid, s.AbsolutePressure, s.RelativePressure, s.OutdoorTemp, s.OutdoorHumid, s.WindDirection, s.WindSpeed, s.WindGust, s.SolarRadiation, s.UV, s.UVI, s.HourlyRain, s.DailyRain, s.WeeklyRain, s.MonthlyRain, s.YearlyRain)
	if err != nil {
		log.Printf("Error begin database %v\n", err.Error())
		return err
	}
	if err = tx.Commit(); err != nil {
		log.Printf("Error commit database %v\n", err.Error())
		return err
	}
	return err
}
