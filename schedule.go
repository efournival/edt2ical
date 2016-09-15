package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/colm2/ical"
)

var (
	isTimeRange    = regexp.MustCompile(`(?i)^(\d{1,2})h(\d{2})\s*-\s*(\d{1,2})h(\d{2})$`)
	isDay          = regexp.MustCompile(`(?i)^(lundi|mardi|mercredi|jeudi|vendredi|samedi|dimanche)$`)
	isDate         = regexp.MustCompile(`^(\d{2})[\.|/](\d{2})[\.|/](\d{2,4})`)
	isGarbage      = regexp.MustCompile(`(?i)(matin|midi|sema|date|conges|\d{2}\s\d{2})`)
	matchGroup     = regexp.MustCompile(`(?i)Gr\.*\s*(\d{1})`)
	matchTimeRange = regexp.MustCompile(`(?i)(\d{1,2})h(\d{2})\s*-\s*(\d{1,2})h(\d{2})`)
	matchLocation  = regexp.MustCompile(`(?mi)(salle\s*:*\s*(.*)|(\w*[\s|\-]*amphi))`)
	timeZone       = "Europe/Paris"
)

type Coords struct {
	X int
	Y int
}

type TimeRange struct {
	start time.Duration
	end   time.Duration
}

type Schedule struct {
	timeRanges map[int]TimeRange // pos X
	days       map[int]string    // pos X
	dates      map[int]string    // pos Y
	entries    map[Coords]string // X, Y
	lineIndex  int
}

func newSchedule() *Schedule {
	return &Schedule{
		timeRanges: make(map[int]TimeRange),
		days:       make(map[int]string),
		dates:      make(map[int]string),
		entries:    make(map[Coords]string),
		lineIndex:  0,
	}
}

func (s *Schedule) parseLine(cols []string) {
	for k, v := range cols {
		sv := strings.TrimSpace(v)

		// Always ignore first column
		if len(sv) > 0 && k > 0 {
			if isTimeRange.MatchString(sv) {
				s.timeRanges[k] = toTimeRange(matchTimeRange.FindAllStringSubmatch(sv, -1)[0])
			} else if isDay.MatchString(sv) {
				s.days[k] = sv
			} else if isDate.MatchString(sv) {
				s.dates[s.lineIndex] = sv
			} else if !isGarbage.MatchString(sv) {
				s.entries[Coords{k, s.lineIndex}] = sv
			}
		}
	}

	s.lineIndex++
}

func (s *Schedule) getDate(x, y int) time.Time {
	rawBaseDate := isDate.FindAllStringSubmatch(mapFindLowerInterval(s.dates, y), -1)

	year := strToInt(rawBaseDate[0][3])

	// XX -> 20XX
	if year < 2000 {
		year += 2000
	} else {
		// Fix for wrong date in original schedule
		year -= 1
	}

	tz, _ := time.LoadLocation(timeZone)
	baseDate := time.Date(year, time.Month(strToInt(rawBaseDate[0][2])), strToInt(rawBaseDate[0][1]), 0, 0, 0, 0, tz)

	day := mapFindLowerInterval(s.days, x)
	return baseDate.Add(time.Duration(getDayOffset(day)*24) * time.Hour)
}

func (s *Schedule) outputCalendar() {
	vcal := ical.NewBasicVCalendar()
	vcal.X_WR_CALNAME = "Emploi du temps M1 Informatique"

	for k, v := range s.entries {
		var ve ical.VEvent
		var tr TimeRange

		d := s.getDate(k.X, k.Y)

		if matchTimeRange.MatchString(v) {
			tr = toTimeRange(matchTimeRange.FindAllStringSubmatch(v, -1)[0])
		} else {
			tr = s.timeRanges[k.X]
		}

		ve.DTSTART = d.Add(tr.start)
		ve.DTEND = d.Add(tr.end)

		if matchLocation.MatchString(v) {
			loc := matchLocation.FindAllStringSubmatch(v, -1)[0]
			ve.LOCATION = loc[2]

			if ve.LOCATION == "" {
				ve.LOCATION = loc[3]
			}
		}

		ve.SUMMARY = strings.Split(v, "\n")[0]
		ve.SUMMARY = strings.TrimSpace(matchTimeRange.ReplaceAllString(ve.SUMMARY, ""))
		ve.SUMMARY = strings.TrimSpace(matchLocation.ReplaceAllString(ve.SUMMARY, ""))

		ve.TZID = timeZone

		vcal.VComponent = append(vcal.VComponent, &ve)
	}

	var b bytes.Buffer

	if err := vcal.Encode(&b); err != nil {
		panic(err)
	}

	fmt.Print(b.String())
}
