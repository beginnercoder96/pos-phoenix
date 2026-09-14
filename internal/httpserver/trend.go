package httpserver

import (
	"fmt"
	"time"

	transactionstore "github.com/mekari/pos-phoenix/internal/transaction"
)

// buildTrend creates a clean, period-aware trend array for the cash-flow chart.
// When period is:
// - "today": 8 distinct 3-hour intervals across today (00:00, 03:00, ..., 21:00)
// - "month": the last 6 months (e.g. Apr, May, Jun, Jul, Aug, Sep)
// - "custom": daily points between customFrom and customTo, capped at the month's days (28, 29, 30, or 31)
// - "week" (default): the last 7 distinct days (e.g. 07 Sep .. 13 Sep)
func buildTrend(entries []transactionstore.Entry, location *time.Location, now time.Time, period string, customRange ...time.Time) []trendPoint {
	nowInLoc := now.In(location)
	todayStart := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), 0, 0, 0, 0, location)

	var points []trendPoint

	switch period {
	case "custom":
		var cStart, cEnd time.Time
		if len(customRange) >= 2 && !customRange[0].IsZero() && !customRange[1].IsZero() {
			cStart = time.Date(customRange[0].In(location).Year(), customRange[0].In(location).Month(), customRange[0].In(location).Day(), 0, 0, 0, 0, location)
			cEnd = time.Date(customRange[1].In(location).Year(), customRange[1].In(location).Month(), customRange[1].In(location).Day(), 0, 0, 0, 0, location)
		} else {
			// default to current month 1st to today
			cStart = time.Date(nowInLoc.Year(), nowInLoc.Month(), 1, 0, 0, 0, 0, location)
			cEnd = todayStart
		}
		if cEnd.Before(cStart) {
			cEnd = cStart
		}
		// Maximum selectable days is 28, 29, 30, or 31 based on the month shown (cStart's month)
		daysInMonth := time.Date(cStart.Year(), cStart.Month()+1, 0, 0, 0, 0, 0, location).Day()
		numDays := int(cEnd.Sub(cStart).Hours()/24) + 1
		if numDays > daysInMonth {
			numDays = daysInMonth
			cEnd = cStart.AddDate(0, 0, numDays-1)
		}
		if numDays < 1 {
			numDays = 1
		}

		points = make([]trendPoint, numDays)
		for i := 0; i < numDays; i++ {
			dStart := cStart.AddDate(0, 0, i)
			dEnd := dStart.AddDate(0, 0, 1)

			points[i].Label = dStart.Format("02 Jan")
			points[i].PeriodRange = dStart.Format("Monday, 02 Jan 2006")

			for _, e := range entries {
				t := e.OccurredAt.In(location)
				if (t.Equal(dStart) || t.After(dStart)) && t.Before(dEnd) {
					if e.Kind == "income" {
						points[i].IncomeCents += e.AmountCents
					} else {
						points[i].ExpenseCents += e.AmountCents
					}
				}
			}
		}

	case "today":
		// 8 intervals of 3 hours each: 00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00
		points = make([]trendPoint, 8)
		for i := 0; i < 8; i++ {
			start := todayStart.Add(time.Duration(i*3) * time.Hour)
			end := start.Add(3 * time.Hour)
			points[i].Label = start.Format("15:04")
			points[i].PeriodRange = fmt.Sprintf("%s - %s", start.Format("15:04"), end.Format("15:04"))

			for _, e := range entries {
				t := e.OccurredAt.In(location)
				if (t.Equal(start) || t.After(start)) && t.Before(end) {
					if e.Kind == "income" {
						points[i].IncomeCents += e.AmountCents
					} else {
						points[i].ExpenseCents += e.AmountCents
					}
				}
			}
		}

	case "month":
		// Last 6 months
		points = make([]trendPoint, 6)
		for i := 0; i < 6; i++ {
			monthOffset := i - 5
			mStart := time.Date(nowInLoc.Year(), nowInLoc.Month()+time.Month(monthOffset), 1, 0, 0, 0, 0, location)
			mEnd := time.Date(mStart.Year(), mStart.Month()+1, 1, 0, 0, 0, 0, location)

			points[i].Label = mStart.Format("Jan")
			points[i].PeriodRange = mStart.Format("January 2006")

			for _, e := range entries {
				t := e.OccurredAt.In(location)
				if (t.Equal(mStart) || t.After(mStart)) && t.Before(mEnd) {
					if e.Kind == "income" {
						points[i].IncomeCents += e.AmountCents
					} else {
						points[i].ExpenseCents += e.AmountCents
					}
				}
			}
		}

	default: // "week" - Last 7 Days
		points = make([]trendPoint, 7)
		for i := 0; i < 7; i++ {
			dayOffset := i - 6
			dStart := todayStart.AddDate(0, 0, dayOffset)
			dEnd := dStart.AddDate(0, 0, 1)

			points[i].Label = dStart.Format("02 Jan")
			points[i].PeriodRange = dStart.Format("Monday, 02 Jan 2006")

			for _, e := range entries {
				t := e.OccurredAt.In(location)
				if (t.Equal(dStart) || t.After(dStart)) && t.Before(dEnd) {
					if e.Kind == "income" {
						points[i].IncomeCents += e.AmountCents
					} else {
						points[i].ExpenseCents += e.AmountCents
					}
				}
			}
		}
	}

	max := int64(1)
	for _, p := range points {
		if p.IncomeCents > max {
			max = p.IncomeCents
		}
		if p.ExpenseCents > max {
			max = p.ExpenseCents
		}
	}

	for i := range points {
		points[i].IncomeHeight = int(points[i].IncomeCents * 100 / max)
		points[i].ExpenseHeight = int(points[i].ExpenseCents * 100 / max)
		if points[i].IncomeCents > 0 && points[i].IncomeHeight < 8 {
			points[i].IncomeHeight = 8
		}
		if points[i].ExpenseCents > 0 && points[i].ExpenseHeight < 8 {
			points[i].ExpenseHeight = 8
		}
	}

	return points
}
