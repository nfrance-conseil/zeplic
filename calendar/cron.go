// Package calendar contains: cron.go - format.go - retention.go
//
// Cron provides the functions to extract the information of JSON crontab
// Returns human readable format
//
package calendar

import (
	"strconv"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/utils"
)

var (
	w = config.LogBook()
)

// Crontab extract the information of JSON crontab
func Crontab(cron string) ([]int, []int, []int, []time.Month, []time.Weekday) {
	minute := utils.Before(cron, " ")
	cron = cron[len(minute)+1:]
	hour := utils.Before(cron, " ")
	cron = cron[len(hour)+1:]
	monthday := utils.Before(cron, " ")
	cron = cron[len(monthday)+1:]
	month := utils.Before(cron, " ")
	weekday := cron[len(month)+1:]

	// Minute
	var cMinute []int
	if minute == "*" {
		for i := 0; i < 60; i++ {
			cMinute = append(cMinute, i)
		}
	} else if strings.Contains(minute, "/") {
		M := utils.After(minute, "/")
		Mint, _ := strconv.Atoi(M)
		var Msum int
		if Mint < 0 || Mint > 59 {
			w.Err("[ERROR > utils/cron.go:42] remember MM{00-59}.")
		} else {
			for Msum < 60 {
				cMinute = append(cMinute, Msum)
				Msum += Mint
			}
		}
	} else {
		for len(minute) > 0 {
			if strings.Contains(minute, ",") {
				M := utils.Before(minute, ",")
				Mint, _ := strconv.Atoi(M)
				if Mint < 0 || Mint > 59 {
					w.Err("[ERROR > utils/cron.go:55] remember MM{00-59}.")
				} else {
					cMinute = append(cMinute, Mint)
					minute = minute[len(M)+1:]
				}
			} else {
				Mint, _ := strconv.Atoi(minute)
				if Mint < 0 || Mint > 59 {
					w.Err("[ERROR > utils/cron.go:63] remember MM{00-59}.")
				} else {
					cMinute = append(cMinute, Mint)
					minute = minute[len(minute):]
				}
			}
		}
	}

	// Hour
	var cHour []int
	if hour == "*" {
		for i := 0; i < 24; i++ {
			cHour = append(cHour, i)
		}
	} else {
		for len(hour) > 0 {
			if strings.Contains(hour, ",") {
				H := utils.Before(hour, ",")
				Hint, _ := strconv.Atoi(H)
				if Hint < 0 || Hint > 23 {
					w.Err("[ERROR > utils/cron.go:84] remember HH{00-23}.")
				} else {
					cHour = append(cHour, Hint)
					hour = hour[len(H)+1:]
				}
			} else {
				Hint, _ := strconv.Atoi(hour)
				if Hint < 0 || Hint > 23 {
					w.Err("[ERROR > utils/cron.go:92] remember HH{00-23}.")
				} else {
					cHour = append(cHour, Hint)
					hour = hour[len(hour):]
				}
			}
		}
	}

	// Monthday
	var cMonthday []int
	if monthday == "*" {
		for i := 1; i < 32; i++ {
			cMonthday = append(cMonthday, i)
		}
	} else if strings.Contains(monthday, "-") {
		fmonthdayString := utils.Before(monthday, "-")
		fmonthday, _ := strconv.Atoi(fmonthdayString)
		lmonthdayString := utils.After(monthday, "-")
		lmonthday, _ := strconv.Atoi(lmonthdayString)
		if fmonthday < 1 || lmonthday > 31 {
			w.Err("[ERROR > utils/cron.go:113] remember dmonth{1-31}.")
		} else {
			for i := fmonthday; i < lmonthday+1; i++ {
				cMonthday = append(cMonthday, i)
			}
		}
	} else {
		monthdayInt, _ := strconv.Atoi(monthday)
		cMonthday = append(cMonthday, monthdayInt)
	}

	// Month
	var cMonth []time.Month
	if month == "*" {
		for i := 1; i < 13; i++ {
			Month := NameMonthInt(i)
			cMonth = append(cMonth, Month)
		}
	} else if strings.Contains(month, "-") {
		fmonthString := utils.Before(month, "-")
		fmonth, _ := strconv.Atoi(fmonthString)
		lmonthString := utils.After(month, "-")
		lmonth, _ := strconv.Atoi(lmonthString)
		if fmonth < 1 || lmonth > 12 {
			w.Err("[ERROR > utils/cron.go:137] remember month{1-12}.")
		} else {
			for i := fmonth; i < lmonth+1; i++ {
				Month := NameMonthInt(i)
				cMonth = append(cMonth, Month)
			}
		}
	} else {
		monthInt, _ := strconv.Atoi(month)
		Month := NameMonthInt(monthInt)
		cMonth = append(cMonth, Month)
	}

	// Weekday
	var cWeekday []time.Weekday
	if weekday == "0-6" || weekday == "1-7" || weekday == "*" {
		for i := 0; i < 7; i++ {
			Weekday := NameWeekInt(i)
			cWeekday = append(cWeekday, Weekday)
		}
	} else if weekday == "1-5" {
		for i := 1; i < 6; i++ {
			Weekday := NameWeekInt(i)
			cWeekday = append(cWeekday, Weekday)
		}
	} else if strings.Contains(weekday, "-") {
		fweekdayString := utils.Before(weekday, "-")
		fweekday, _ := strconv.Atoi(fweekdayString)
		lweekdayString := utils.After(weekday, "-")
		lweekday, _ := strconv.Atoi(lweekdayString)
		if fweekday == 0 && lweekday == 7 {
			w.Err("[ERROR > utils/cron.go:168] remember dweek{0-6/1-7}.")
		} else if fweekday < 0 || lweekday > 7 {
			w.Err("[ERROR > utils/cron.go:170] remember dweek{0-6/1-7}.")
		} else {
			if lweekday == 7 {
				lweekday = 0
			}
			for i := fweekday; i < lweekday+1; i++ {
				Weekday := NameWeekInt(i)
				cWeekday = append(cWeekday, Weekday)
			}
		}
	} else {
		weekdayInt, _ := strconv.Atoi(weekday)
		Weekday := NameWeekInt(weekdayInt)
		cWeekday = append(cWeekday, Weekday)
	}

	return cMinute, cHour, cMonthday, cMonth, cWeekday
}
