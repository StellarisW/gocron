package regex

import (
	"errors"
	"regexp"
)

func getRegexp(pattern string) (regex *regexp.Regexp, err error) {
	if regex, err = regexp.Compile(pattern); err != nil {
		err = errors.New(`regexp.Compile failed`)
		return
	}
	return
}

func IsMatch(pattern string, src []byte) bool {
	if r, err := getRegexp(pattern); err == nil {
		return r.Match(src)
	}
	return false
}

func MatchString(pattern string, src string) ([]string, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.FindStringSubmatch(src), nil
	} else {
		return nil, err
	}
}

func IsMatchString(pattern string, src string) bool {
	return IsMatch(pattern, []byte(src))
}
