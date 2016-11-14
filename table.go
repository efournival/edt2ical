package main

import (
	"log"
	"strings"
	"time"
)

type Table struct {
	timeRanges map[int]TimeRange
	days       map[int]string
	dates      map[int]string
	entries    map[Coords]string
}

func (t *Table) splitEntry(x, y int, str string) {
	// Workaround for "PARTIEL:" and the name of the course on the next line
	str = strings.Replace(str, ":\n", " ", -1)

	if matchGroupLocation.MatchString(str) {
		gl := matchGroupLocation.FindAllStringSubmatch(str, -1)

		lines := strings.Split(str, "\n")
		firstLine := strings.TrimSpace(lines[0])
		firstLine = matchGroup.ReplaceAllString(firstLine, "")

		for i := 0; i < len(gl); i++ {
			t.entries[Coords{x, y, i}] = firstLine + "\n" + gl[i][0]

			if debug {
				log.Printf("(%d:%d:%d) Splitting group: '%s'\n", x, y, i, one(t.entries[Coords{x, y, i}]))
			}
		}

		return
	}

	if matchTERLocation.MatchString(str) {
		gl := matchTERLocation.FindAllStringSubmatch(str, -1)

		for i := 0; i < len(gl); i++ {
			t.entries[Coords{x, y, i}] = gl[i][0]

			if debug {
				log.Printf("(%d:%d:%d) Splitting TER: '%s'\n", x, y, i, one(t.entries[Coords{x, y, i}]))
			}
		}

		return
	}

	t.entries[Coords{x, y, 0}] = str
}

func (t *Table) getDate(x, y int) time.Time {
	rawBaseDate := isDate.FindAllStringSubmatch(mapFindLowerIntervalS(t.dates, y), -1)

	year := strToInt(rawBaseDate[0][3])

	if year < 2000 {
		year += 2000
	}

	tz, _ := time.LoadLocation(timeZone)
	baseDate := time.Date(year, time.Month(strToInt(rawBaseDate[0][2])), strToInt(rawBaseDate[0][1]), 0, 0, 0, 0, tz)

	day := mapFindLowerIntervalS(t.days, x)
	return baseDate.Add(time.Duration(getDayOffset(day)*24) * time.Hour)
}
