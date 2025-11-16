package repository

import "tltui/src/domain"

func fetchAllProjects() []domain.Project {
	return []domain.Project{
		{ID: 1, Name: "Campoint", OdooID: 102},
		{ID: 2, Name: "Campoint Overtime", OdooID: 102},
		{ID: 3, Name: "Arnia", OdooID: 40},
		{ID: 4, Name: "DSWISS", OdooID: 192},
		{ID: 5, Name: "Standford", OdooID: 287},
		{ID: 6, Name: "BCS AG", OdooID: 293},
	}
}

func fetchAllWorkhourDetails() []domain.WorkhourDetails {
	return []domain.WorkhourDetails{
		{ID: 1, Name: "Development", ShortName: "🔧", IsWork: true},
		{ID: 2, Name: "Leave", ShortName: "🏖️", IsWork: false},
		{ID: 3, Name: "National Day", ShortName: "🇷🇴", IsWork: false},
	}
}
