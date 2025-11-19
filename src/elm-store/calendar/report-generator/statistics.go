package report_generator

import (
	"tltui/src/domain"
)

// WorkhourStats contains aggregated statistics for workhours
type WorkhourStats struct {
	TotalHours           float64
	TotalDays            int
	AveragePerDay        float64
	ProjectHours         map[string]float64            // project name -> hours
	ActivityHours        map[string]float64            // activity name -> hours
	ProjectActivityHours map[string]map[string]float64 // project name -> activity name -> hours
	DailyBreakdown       map[string][]WorkhourEntry    // date -> list of entries
}

// WorkhourEntry represents a single workhour entry
type WorkhourEntry struct {
	ProjectName  string
	ActivityName string
	Hours        float64
}

// CalculateWorkhourStats calculates comprehensive statistics from workhours
func CalculateWorkhourStats(
	workhours []domain.Workhour,
	detailsMap map[int]domain.WorkhourDetails,
	projectsMap map[int]domain.Project,
) WorkhourStats {
	stats := WorkhourStats{
		ProjectHours:         make(map[string]float64),
		ActivityHours:        make(map[string]float64),
		ProjectActivityHours: make(map[string]map[string]float64),
		DailyBreakdown:       make(map[string][]WorkhourEntry),
	}

	daysWorked := make(map[string]bool)

	for _, wh := range workhours {
		stats.TotalHours += wh.Hours

		dateStr := wh.Date.Format("02-Jan-2006")
		daysWorked[dateStr] = true

		var projectName, activityName string

		if project, ok := projectsMap[wh.ProjectID]; ok {
			projectName = project.Name
			stats.ProjectHours[projectName] += wh.Hours
		}

		if details, ok := detailsMap[wh.DetailsID]; ok {
			activityName = details.Name
			stats.ActivityHours[activityName] += wh.Hours
		}

		if projectName != "" && activityName != "" {
			if stats.ProjectActivityHours[projectName] == nil {
				stats.ProjectActivityHours[projectName] = make(map[string]float64)
			}
			stats.ProjectActivityHours[projectName][activityName] += wh.Hours
		}

		entry := WorkhourEntry{
			Hours:        wh.Hours,
			ProjectName:  projectName,
			ActivityName: activityName,
		}
		stats.DailyBreakdown[dateStr] = append(stats.DailyBreakdown[dateStr], entry)
	}

	stats.TotalDays = len(daysWorked)
	if stats.TotalDays > 0 {
		stats.AveragePerDay = stats.TotalHours / float64(stats.TotalDays)
	}

	return stats
}
