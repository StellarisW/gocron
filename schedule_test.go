package gocron

import (
	"testing"
	"time"
)

func getTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse("Mon Jan 2 15:04 2006", value)
	if err != nil {
		t, err = time.Parse("Mon Jan 2 15:04:05 2006", value)
		if err != nil {
			t, err = time.Parse("Mon Jan 2 15:04 2006 MST", value)
			if err != nil {
				panic(err)
			}
		}
	}
	return t
}

func TestNext(t *testing.T) {
	runs := []struct {
		time, pattern string
		expected      string
	}{
		// Simple cases
		{"Tue Jul 19 14:45 2022", "0 * */15 * * *", "Tue Jul 19 15:00 2022"},
		{"Tue Jul 19 14:59 2022", "0 * */15 * * *", "Tue Jul 19 15:00 2022"},
		{"Tue Jul 19 14:59:59 2022", "0 * */15 * * *", "Tue Jul 19 15:00 2022"},

		// Wrap around hours
		{"Tue Jul 19 15:45 2022", "0 20-35/15 * * * *", "Tue Jul 19 16:20 2022"},

		// Wrap around days
		{"Tue Jul 19 23:46 2022", "0 * */15 * * *", "Wed Jul 20 00:00 2022"},
		{"Tue Jul 19 23:45 2022", "0 20-35/15 * * * *", "Wed Jul 20 00:20 2022"},
		{"Tue Jul 19 23:35:51 2022", "15/35 20-35/15 * * * *", "Wed Jul 20 00:20:15 2022"},
		{"Tue Jul 19 23:35:51 2022", "15/35 20-35/15 1/2 * * *", "Wed Jul 20 01:20:15 2022"},
		{"Tue Jul 19 23:35:51 2022", "15/35 20-35/15 10-12 * * *", "Wed Jul 20 10:20:15 2022"},

		{"Tue Jul 19 23:35:51 2022", "15/35 20-35/15 1/2 */2 * *", "Thu Jul 21 01:20:15 2022"},
		{"Tue Jul 19 23:35:51 2022", "15/35 20-35/15 * 9-20 * *", "Wed Jul 20 00:20:15 2022"},
		{"Tue Jul 19 23:35:51 2022", "15/35 20-35/15 * 9-20 Jul *", "Wed Jul 20 00:20:15 2022"},

		// Wrap around months
		{"Mon Jul 9 23:35 2012", "0 0 0 9 Apr-Oct ?", "Thu Aug 9 00:00 2012"},
		{"Mon Jul 9 23:35 2012", "0 0 0 */5 Apr,Aug,Oct Mon", "Mon Aug 6 00:00 2012"},
		{"Mon Jul 9 23:35 2012", "0 0 0 */5 Oct Mon", "Mon Oct 1 00:00 2012"},

		// Wrap around years
		{"Mon Jul 9 23:35 2012", "0 0 0 * Feb Mon", "Mon Feb 4 00:00 2013"},
		{"Mon Jul 9 23:35 2012", "0 0 0 * Feb Mon/2", "Fri Feb 1 00:00 2013"},

		// Wrap around minute, hour, day, month, and year
		{"Mon Dec 31 23:59:45 2012", "0 * * * * *", "Tue Jan 1 00:00:00 2013"},

		// Leap year
		{"Mon Jul 9 23:35 2012", "0 0 0 29 Feb ?", "Mon Feb 29 00:00 2016"},

		// Daylight savings time
		// TODO: 待解决bug
		{"Sun Mar 12 00:00 2022 EST", "0 30 2 11 Mar ?", "Mon Mar 11 02:30 2023 EST"},

		// Unsatisfiable
		// TODO: 待解决,为什么没有返回zero time
		{"Mon Jul 9 15:45 2012", "0 * 20-35 * * *", ""},
		{"Mon Jul 9 23:35 2012", "0 0 0 30 Feb ?", ""},
		{"Mon Jul 9 23:35 2012", "0 0 0 31 Apr ?", ""},
	}

	for _, c := range runs {
		schedule, err := newSchedule(c.pattern)
		if err != nil {
			t.Errorf("%s, \"%s\": got err:%v", c.time, c.pattern, err.Error())
			continue
		}
		actual := schedule.Next(getTime(c.time))
		expected := getTime(c.expected)
		if actual != expected {
			t.Errorf("%s, \"%s\": (expected) %v != %v (actual)", c.time, c.pattern, expected, actual)
		}
	}
}
