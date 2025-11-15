package models

func FetchAllProjects() []Project {
	return []Project{
		{ID: 1, Name: "Project A", OdooID: 101},
		{ID: 2, Name: "Project B", OdooID: 102},
		{ID: 3, Name: "Internal Tools", OdooID: 103},
		{ID: 4, Name: "Client Website", OdooID: 104},
		{ID: 5, Name: "Mobile App", OdooID: 105},
		{ID: 6, Name: "E-Commerce Platform", OdooID: 106},
		{ID: 7, Name: "CRM System", OdooID: 107},
		{ID: 8, Name: "Analytics Dashboard", OdooID: 108},
		{ID: 9, Name: "Payment Gateway", OdooID: 109},
		{ID: 10, Name: "Inventory Management", OdooID: 110},
		{ID: 11, Name: "HR Portal", OdooID: 111},
		{ID: 12, Name: "Customer Support Tool", OdooID: 112},
		{ID: 13, Name: "Marketing Automation", OdooID: 113},
		{ID: 14, Name: "API Gateway", OdooID: 114},
		{ID: 15, Name: "DevOps Pipeline", OdooID: 115},
		{ID: 16, Name: "Security Audit", OdooID: 116},
		{ID: 17, Name: "Database Migration", OdooID: 117},
		{ID: 18, Name: "Cloud Infrastructure", OdooID: 118},
		{ID: 19, Name: "Machine Learning Model", OdooID: 119},
		{ID: 20, Name: "IoT Platform", OdooID: 120},
	}
}

func FetchAllWorkhourDetails() []WorkhourDetails {
	return []WorkhourDetails{
		{ID: 1, Name: "Regular Work", ShortName: "Work", IsWork: true},
		{ID: 2, Name: "Overtime", ShortName: "OT", IsWork: true},
		{ID: 3, Name: "Vacation", ShortName: "Vac", IsWork: false},
		{ID: 4, Name: "Sick Leave", ShortName: "Sick", IsWork: false},
		{ID: 5, Name: "Training", ShortName: "Train", IsWork: true},
		{ID: 6, Name: "Meeting", ShortName: "Meet", IsWork: true},
		{ID: 7, Name: "Break", ShortName: "Break", IsWork: false},
		{ID: 8, Name: "Holiday", ShortName: "Holiday", IsWork: false},
		{ID: 9, Name: "Remote Work", ShortName: "Remote", IsWork: true},
		{ID: 10, Name: "On-Call", ShortName: "OnCall", IsWork: true},
	}
}
