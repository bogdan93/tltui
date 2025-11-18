package repository

import "tltui/src/domain"

func fetchAllProjects() []domain.Project {
	return []domain.Project{
		{ID: 1, Name: "Campoint", OdooID: 102},
		{ID: 3, Name: "Arnia", OdooID: 40},
	}
}

func fetchAllWorkhourDetails() []domain.WorkhourDetails {
	return []domain.WorkhourDetails{
		{ID: 1, Name: "Development", ShortName: "🔧", IsWork: true},
		{ID: 2, Name: "Development Overtime", ShortName: "🕐", IsWork: true},
		{ID: 3, Name: "Leave", ShortName: "🏖️", IsWork: false},
		{ID: 4, Name: "National Day", ShortName: "🇷🇴", IsWork: false},
	}
}
