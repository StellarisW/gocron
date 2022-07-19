package gocron

import (
	"errors"
	"gocron/regex"
	"strconv"
	"strings"
	"time"
)

// Schedule 存储entry的具体执行时间
type Schedule struct {
	everySeconds int64            // 时间间隔.
	pattern      string           // 原始表达式.
	secondMap    map[int]struct{} // Job 在每一分钟的这些秒运行
	minuteMap    map[int]struct{} // Job 在每一小时的这些分钟运行
	hourMap      map[int]struct{} // Job 在每一天的这些小时运行
	dayMap       map[int]struct{} // Job 在每个月的这些天运行
	weekMap      map[int]struct{} // Job 在每个月的这些星期运行
	monthMap     map[int]struct{} // Job 在每一年的这些月运行
}

const (
	// cron的正则表达式,包含6个部分
	regexForCron           = `^([\-/\d\*\?,]+)\s+([\-/\d\*\?,]+)\s+([\-/\d\*\?,]+)\s+([\-/\d\*\?,]+)\s+([\-/\d\*\?,A-Za-z]+)\s+([\-/\d\*\?,A-Za-z]+)$`
	patternItemTypeUnknown = iota
	patternItemTypeWeek
	patternItemTypeMonth
)

var (
	// 关于月份简称的映射表
	monthShortNameMap = map[string]int{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}
	// 关于月份全称的映射表
	monthFullNameMap = map[string]int{
		"january":   1,
		"february":  2,
		"march":     3,
		"april":     4,
		"may":       5,
		"june":      6,
		"july":      7,
		"august":    8,
		"september": 9,
		"october":   10,
		"november":  11,
		"december":  12,
	}
	// 关于星期简称的映射表
	weekShortNameMap = map[string]int{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}
	// 关于星期全称的映射表
	weekFullNameMap = map[string]int{
		"sunday":    0,
		"monday":    1,
		"tuesday":   2,
		"wednesday": 3,
		"thursday":  4,
		"friday":    5,
		"saturday":  6,
	}
)

func newSchedule(pattern string) (*Schedule, error) {
	if match, _ := regex.MatchString(regexForCron, pattern); len(match) == 7 {
		schedule := &Schedule{
			everySeconds: 0,
			pattern:      pattern,
		}
		if m, err := parsePatternItem(match[1], 0, 59, false); err != nil {
			return nil, err
		} else {
			schedule.secondMap = m
		}
		if m, err := parsePatternItem(match[2], 0, 59, false); err != nil {
			return nil, err
		} else {
			schedule.minuteMap = m
		}
		if m, err := parsePatternItem(match[3], 0, 23, false); err != nil {
			return nil, err
		} else {
			schedule.hourMap = m
		}
		if m, err := parsePatternItem(match[4], 1, 31, true); err != nil {
			return nil, err
		} else {
			schedule.dayMap = m
		}
		if m, err := parsePatternItem(match[5], 1, 12, false); err != nil {
			return nil, err
		} else {
			schedule.monthMap = m
		}
		if m, err := parsePatternItem(match[6], 0, 6, true); err != nil {
			return nil, err
		} else {
			schedule.weekMap = m
		}
		return schedule, nil
	}
	return nil, errors.New("invalid pattern")
}

// parsePatternItem 解析表达式中的单个部分
func parsePatternItem(item string, min int, max int, allowQuestionMark bool) (map[int]struct{}, error) {
	m := make(map[int]struct{}, max-min+1)
	if item == "*" || (allowQuestionMark && item == "?") {
		for i := min; i <= max; i++ {
			m[i] = struct{}{}
		}
		return m, nil
	}
	// Like: MON,FRI
	for _, itemElem := range strings.Split(item, ",") {
		var (
			interval      = 1
			intervalArray = strings.Split(itemElem, "/")
		)
		if len(intervalArray) == 2 {
			if number, err := strconv.Atoi(intervalArray[1]); err != nil {
				return nil, errors.New("invalid pattern item")
			} else {
				interval = number
			}
		}
		var (
			rangeMin   = min
			rangeMax   = max
			itemType   = patternItemTypeUnknown
			rangeArray = strings.Split(intervalArray[0], "-") // Like: 1-30, JAN-DEC
		)
		switch max {
		case 6:
			// 检测week字段
			itemType = patternItemTypeWeek

		case 12:
			// 检测month字段
			itemType = patternItemTypeMonth
		}
		// Eg: */5
		if rangeArray[0] != "*" {
			if number, err := parsePatternItemValue(rangeArray[0], itemType); err != nil {
				return nil, errors.New("invalid pattern item")
			} else {
				rangeMin = number
				if len(intervalArray) == 1 {
					rangeMax = number
				}
			}
		}
		if len(rangeArray) == 2 {
			if number, err := parsePatternItemValue(rangeArray[1], itemType); err != nil {
				return nil, errors.New("invalid pattern item")
			} else {
				rangeMax = number
			}
		}
		for i := rangeMin; i <= rangeMax; i += interval {
			m[i] = struct{}{}
		}
	}
	return m, nil
}

// parsePatternItemValue 解析表达式中的单个部分的具体值
func parsePatternItemValue(value string, itemType int) (int, error) {
	if regex.IsMatchString(`^\d+$`, value) {
		// 纯数字
		if number, err := strconv.Atoi(value); err == nil {
			return number, nil
		}
	} else {
		// 检测是否包含字母
		// 将月份,星期进行对数字的映射
		switch itemType {
		case patternItemTypeWeek:
			if number, ok := weekShortNameMap[strings.ToLower(value)]; ok {
				return number, nil
			}
			if number, ok := weekFullNameMap[strings.ToLower(value)]; ok {
				return number, nil
			}
		case patternItemTypeMonth:
			if number, ok := monthShortNameMap[strings.ToLower(value)]; ok {
				return number, nil
			}
			if number, ok := monthFullNameMap[strings.ToLower(value)]; ok {
				return number, nil
			}
		}
	}
	return 0, errors.New("invalid pattern value")
}

// Next 返回下一次工作函数运行的时间,找不到返回zero time
func (s *Schedule) Next(t time.Time) time.Time {
	if s.everySeconds != 0 {
		return t.Add(time.Duration(s.everySeconds) * time.Second)
	}

	// 下面参考了一些代码来做
	// Reference: https://github.com/robfig/cron/blob/master/spec.go#L82

	// 提前一点点时间运行,使其更精确
	t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)
	var (
		loc       = t.Location()
		added     = false
		yearLimit = t.Year() + 5
	)

WRAP:
	if t.Year() > yearLimit {
		return t
	}

	for !s.match(s.monthMap, int(t.Month())) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
		t = t.AddDate(0, 1, 0)
		// need recheck
		if t.Month() == time.January {
			goto WRAP
		}
	}

	for !s.dayMatches(t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		}
		t = t.AddDate(0, 0, 1)

		if t.Hour() != 0 {
			if t.Hour() > 12 {
				t = t.Add(time.Duration(24-t.Hour()) * time.Hour)
			} else {
				t = t.Add(time.Duration(-t.Hour()) * time.Hour)
			}
		}
		if t.Day() == 1 {
			goto WRAP
		}
	}
	for !s.match(s.hourMap, t.Hour()) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
		}
		t = t.Add(time.Hour)
		// need recheck
		if t.Hour() == 0 {
			goto WRAP
		}
	}
	for !s.match(s.minuteMap, t.Minute()) {
		if !added {
			added = true
			t = t.Truncate(time.Minute)
		}
		t = t.Add(1 * time.Minute)

		if t.Minute() == 0 {
			goto WRAP
		}
	}
	for !s.match(s.secondMap, t.Second()) {
		if !added {
			added = true
			t = t.Truncate(time.Second)
		}
		t = t.Add(1 * time.Second)
		if t.Second() == 0 {
			goto WRAP
		}
	}
	return t.In(loc)
}

// dayMatches 进行月份和星期当中的天数匹配
func (s *Schedule) dayMatches(t time.Time) bool {
	_, ok1 := s.dayMap[t.Day()]
	_, ok2 := s.weekMap[int(t.Weekday())]
	return ok1 && ok2
}

// 匹配字段值
func (s *Schedule) match(m map[int]struct{}, key int) bool {
	_, ok := m[key]
	return ok
}
