package test

import (
	"time"

	"github.com/nfrance-conseil/zeplic/calendar"
	"testing"
)

func TestCrontab(t *testing.T) {
	cron := "0 12 1 7 1-5"
	cMinute, cHour, cMonthday, cMonth, cWeekday := calendar.Crontab(cron)
	if cMinute[0] != 0 || cHour[0] != 12 || cMonthday[0] != 1 || cMonth[0] != time.July || cWeekday[4] != time.Friday {
		t.Errorf("Crontab() test failed!")
	}
}
