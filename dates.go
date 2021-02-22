package mys_sap_client

import (
	"fmt"
	"time"
)

//Converts a sap string datetime in format yyyyMMddHHmmss to a go time
func FromSAPDateTime(date, t string) time.Time {
	const layout = "20060102150405"
	parsed, err := time.Parse(layout, date+t)
	if err != nil {
		return time.Now()
	}
	return parsed
}

//Converts a sap string date in format yyyyMMdd to a go time
func FromSAPDate(date string) time.Time {
	const layout = "20060102"
	parsed, err := time.Parse(layout, date)
	if err != nil {
		return time.Now()
	}
	return parsed
}

var location *time.Location

func SetTimeLocation(name string) {
	l, err := time.LoadLocation(name)
	if err != nil {
		fmt.Printf("error setting SAP time location \n")
		return
	}
	location = l

}

func localeTime(t time.Time) time.Time {
	if location == nil {
		return t
	}
	return t.In(location)
}

//Formats a time into a SAP date string in format yyyyMMdd
func ToSAPDate(t time.Time) string {
	const layout = "20060102"
	return localeTime(t).Format(layout)
}

//Formats a time into a SAP time string in format HHmmss
func ToSapTime(t time.Time) string {
	const layout = "150405"
	return localeTime(t).Format(layout)
}

//Formats a time into a SAP datetime string in format yyyyMMddHHmmss
func ToSAPDateTime(t time.Time) string {
	const layout = "20060102150405"
	return localeTime(t).Format(layout)
}
