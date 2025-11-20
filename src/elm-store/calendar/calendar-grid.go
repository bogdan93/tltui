package calendar

import "time"

// getCalendarGrid returns a 6x7 grid of dates for the current view month
// Starting from Monday and including days from previous/next months as needed
func (m CalendarModel) getCalendarGrid() [6][7]time.Time {
	var grid [6][7]time.Time

	firstDay := time.Date(m.ViewYear, time.Month(m.ViewMonth), 1, 0, 0, 0, 0, time.Local)

	weekday := int(firstDay.Weekday()) // 0 = Sunday, 1 = Monday, etc.
	daysFromMonday := (weekday + 6) % 7
	startDate := firstDay.AddDate(0, 0, -daysFromMonday)

	currentDate := startDate
	for week := range 6 {
		for day := range 7 {
			grid[week][day] = currentDate
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	return grid
}

// isSameDay checks if two dates represent the same calendar day
func (m CalendarModel) isSameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// isDateInVisibleGrid checks if a date is visible in the current calendar grid
func (m CalendarModel) isDateInVisibleGrid(date time.Time) bool {
	grid := m.getCalendarGrid()

	normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

	for week := range 6 {
		for day := range 7 {
			gridDate := grid[week][day]
			if m.isSameDay(normalizedDate, gridDate) {
				return true
			}
		}
	}

	return false
}

// ResetToCurrentMonth resets the calendar view to the current month
func (m *CalendarModel) ResetToCurrentMonth() {
	now := time.Now()
	m.SelectedDate = now
	m.ViewMonth = int(now.Month())
	m.ViewYear = now.Year()
}
