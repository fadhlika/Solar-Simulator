package main

import "time"

type solardata struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created"`
	Voltage float64   `json:"voltage"`
	Current float64   `json:"current"`
	Temp1   float64   `json:"temp1"`
	Temp2   float64   `json:"temp2"`
	Lum1    float64   `json:"lum1"`
	Lum2    float64   `json:"lum2"`
	Deleted bool      `json:"deleted"`
}

type solardebug struct {
	ID      int
	Created time.Time
	Message string
	Deleted bool
}

type awsdata struct {
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
	YearlyRain       float64   `json:"yearlyrain"'`
	Deleted          bool      `json:"deleted"`
}
