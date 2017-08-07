// Package tools contains: calendar.go - cron.go - retention.go - substrings.go
//
// Format provides the functions to analyze the name, the number and the lenght of months
//
package tools

import (
	"strings"
	"time"
)

// LengthMonth specifies the length of the month in function of days
func LengthMonth(year int, month time.Month) (bool, string) {
	var leap bool
	list := []int{2020, 2024, 2028, 2032, 2036, 2040}
	for i := 0; i < len(list); i++ {
		if year == list[i] {
			leap = true
			break
		} else {
			continue
		}
	}
	var monthDiff string
	switch month {
	case time.January:
		monthDiff = "long"
	case time.February:
		monthDiff = "long"
	case time.March:
		if leap == true {
			monthDiff = "leap"
		} else {
			monthDiff = "notleap"
		}
	case time.April:
		monthDiff = "long"
	case time.May:
		monthDiff = "short"
	case time.June:
		monthDiff = "long"
	case time.July:
		monthDiff = "short"
	case time.August:
		monthDiff = "long"
	case time.September:
		monthDiff = "long"
	case time.October:
		monthDiff = "short"
	case time.November:
		monthDiff = "long"
	case time.December:
		monthDiff = "short"
	}
	return leap, monthDiff
}

// NameMonth returns the name of the month
func NameMonth(month string) time.Month {
	var Month time.Month
	switch month {
	case "January":
		Month = time.January
	case "February":
		Month = time.February
	case "March":
		Month = time.March
	case "April":
		Month = time.April
	case "May":
		Month = time.May
	case "June":
		Month = time.June
	case "July":
		Month = time.July
	case "August":
		Month = time.August
	case "September":
		Month = time.September
	case "October":
		Month = time.October
	case "November":
		Month = time.November
	case "December":
		Month = time.December
	}
	return Month
}

// NameMonthZero returns the name of the month
func NameMonthZero(month string) time.Month {
	var Month time.Month
	switch month {
	case "01":
		Month = time.January
	case "02":
		Month = time.February
	case "03":
		Month = time.March
	case "04":
		Month = time.April
	case "05":
		Month = time.May
	case "06":
		Month = time.June
	case "07":
		Month = time.July
	case "08":
		Month = time.August
	case "09":
		Month = time.September
	case "10":
		Month = time.October
	case "11":
		Month = time.November
	case "12":
		Month = time.December
	}
	return Month
}

// NameMonthInt returns the name of the month
func NameMonthInt(month int) time.Month {
	var Month time.Month
	switch month {
	case 1:
		Month = time.January
	case 2:
		Month = time.February
	case 3:
		Month = time.March
	case 4:
		Month = time.April
	case 5:
		Month = time.May
	case 6:
		Month = time.June
	case 7:
		Month = time.July
	case 8:
		Month = time.August
	case 9:
		Month = time.September
	case 10:
		Month = time.October
	case 11:
		Month = time.November
	case 12:
		Month = time.December
	}
	return Month
}

// NameWeek returns the name of the weekday
func NameWeek(weekday string) time.Weekday {
	var Weekday time.Weekday
	switch weekday {
	case "Sunday":
		Weekday = time.Sunday
	case "Monday":
		Weekday = time.Monday
	case "Tuesday":
		Weekday = time.Tuesday
	case "Wednesday":
		Weekday = time.Wednesday
	case "Thursday":
		Weekday = time.Thursday
	case "Friday":
		Weekday = time.Friday
	case "Saturday":
		Weekday = time.Saturday
	}
	return Weekday
}

// NameWeekInt returns the name of the weekday
func NameWeekInt(weekday int) time.Weekday {
	var Weekday time.Weekday
	switch weekday {
	case 0:
		Weekday = time.Sunday
	case 1:
		Weekday = time.Monday
	case 2:
		Weekday = time.Tuesday
	case 3:
		Weekday = time.Wednesday
	case 4:
		Weekday = time.Thursday
	case 5:
		Weekday = time.Friday
	case 6:
		Weekday = time.Saturday
	}
	return Weekday
}

// NumberMonth replaces the name of the month by the number
func NumberMonth(SnapshotName string) string {
	if strings.Contains(SnapshotName, "January") {
		SnapshotName = strings.Replace(SnapshotName, "January", "01", -1)
	} else if strings.Contains(SnapshotName, "February") {
		SnapshotName = strings.Replace(SnapshotName, "February", "02", -1)
	} else if strings.Contains(SnapshotName, "March") {
		SnapshotName = strings.Replace(SnapshotName, "March", "03", -1)
	} else if strings.Contains(SnapshotName, "April") {
		SnapshotName = strings.Replace(SnapshotName, "April", "04", -1)
	} else if strings.Contains(SnapshotName, "May") {
		SnapshotName = strings.Replace(SnapshotName, "May", "05", -1)
	} else if strings.Contains(SnapshotName, "June") {
		SnapshotName = strings.Replace(SnapshotName, "June", "06", -1)
	} else if strings.Contains(SnapshotName, "July") {
		SnapshotName = strings.Replace(SnapshotName, "July", "07", -1)
	} else if strings.Contains(SnapshotName, "August") {
		SnapshotName = strings.Replace(SnapshotName, "August", "08", -1)
	} else if strings.Contains(SnapshotName, "September") {
		SnapshotName = strings.Replace(SnapshotName, "September", "09", -1)
	} else if strings.Contains(SnapshotName, "October") {
		SnapshotName = strings.Replace(SnapshotName, "October", "10", -1)
	} else if strings.Contains(SnapshotName, "November") {
		SnapshotName = strings.Replace(SnapshotName, "November", "11", -1)
	} else if strings.Contains(SnapshotName, "December") {
		SnapshotName = strings.Replace(SnapshotName, "December", "12", -1)
	}
	return SnapshotName
}

// NumberMonthReverse replaces the number of the month by the name
func NumberMonthReverse(SnapshotName string) string {
	if strings.Contains(SnapshotName, "-01-") {
		SnapshotName = strings.Replace(SnapshotName, "-01-", "-Janvier-", -1)
	} else if strings.Contains(SnapshotName, "-02-") {
		SnapshotName = strings.Replace(SnapshotName, "-02-", "-February-", -1)
	} else if strings.Contains(SnapshotName, "-03-") {
		SnapshotName = strings.Replace(SnapshotName, "-03-", "-March-", -1)
	} else if strings.Contains(SnapshotName, "-04-") {
		SnapshotName = strings.Replace(SnapshotName, "-04-", "-April-", -1)
	} else if strings.Contains(SnapshotName, "-05-") {
		SnapshotName = strings.Replace(SnapshotName, "-05-", "-May-", -1)
	} else if strings.Contains(SnapshotName, "-06-") {
		SnapshotName = strings.Replace(SnapshotName, "-06-", "-June-", -1)
	} else if strings.Contains(SnapshotName, "-07-") {
		SnapshotName = strings.Replace(SnapshotName, "-07-", "-July-", -1)
	} else if strings.Contains(SnapshotName, "-08-") {
		SnapshotName = strings.Replace(SnapshotName, "-08-", "-August-", -1)
	} else if strings.Contains(SnapshotName, "-09-") {
		SnapshotName = strings.Replace(SnapshotName, "-09-", "-September-", -1)
	} else if strings.Contains(SnapshotName, "-10-") {
		SnapshotName = strings.Replace(SnapshotName, "-10-", "-October-", -1)
	} else if strings.Contains(SnapshotName, "-11-") {
		SnapshotName = strings.Replace(SnapshotName, "-11-", "-November-", -1)
	} else if strings.Contains(SnapshotName, "-12-") {
		SnapshotName = strings.Replace(SnapshotName, "-12-", "-December-", -1)
	}
	return SnapshotName
}
