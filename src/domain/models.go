package domain

import "time"

type Project struct {
	ID     int
	Name   string
	OdooID int
}

type WorkhourDetails struct {
	ID        int
	Name      string
	ShortName string
	IsWork    bool
}

type Workhour struct {
	ID        int
	Date      time.Time
	DetailsID int
	ProjectID int
	Hours     float64
}
