package main

import (
	"strconv"
	"strings"
	"time"
)

func strToInt(str string) int {
	val, err := strconv.Atoi(str)

	if err != nil {
		panic(err)
	}

	return val
}

func strToDuration(str string) time.Duration {
	return time.Duration(strToInt(str))
}

func mapFindLowerInterval(m map[int]string, l int) string {
	last := ""

	for i := 0; i < 100; i++ {
		if v, ok := m[i]; ok {
			last = v
		}

		if i >= l {
			return last
		}
	}

	return last
}

func getDayOffset(day string) int {
	switch strings.ToLower(day) {
	case "mardi":
		return 1
	case "mercredi":
		return 2
	case "jeudi":
		return 3
	case "vendredi":
		return 4
	case "samedi":
		return 5
	case "dimanche":
		return 6
	}

	return 0
}

func toTimeRange(tr []string) TimeRange {
	return TimeRange{strToDuration(tr[1])*time.Hour + strToDuration(tr[2])*time.Minute,
		strToDuration(tr[3])*time.Hour + strToDuration(tr[4])*time.Minute}
}

func isWrongLine(str string) bool {
	str = strings.Split(str, "\n")[0]
	str = matchGroup.ReplaceAllString(str, "")
	str = matchTimeRange.ReplaceAllString(str, "")
	str = matchLocation.ReplaceAllString(str, "")
	str = strings.Split(str, ":")[0]
	str = strings.Split(str, "-")[0]

	return len(strings.TrimSpace(str)) == 0
}
