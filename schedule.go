package main

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/colm2/ical"
)

const (
	timeZone     = "Europe/Paris"
	calendarName = "Emploi du temps M1 Informatique"
	uidHost      = "@edt-m1info.edgar-fournival.fr"
	uidDTformat  = "20060102T150405"
)

var (
	isTimeRange        = regexp.MustCompile(`(?i)^(\d{1,2})\s?h\s?(\d{2})\s*[\-à]\s*(\d{1,2})\s?h\s?(\d{2})$`)
	isDay              = regexp.MustCompile(`(?i)^(lundi|mardi|mercredi|jeudi|vendredi|samedi|dimanche)$`)
	isDate             = regexp.MustCompile(`^(\d{2})[\.|/](\d{2})[\.|/](\d{2,4})`)
	isGarbage          = regexp.MustCompile(`(?i)(matin|midi|semaine|date|conges|page|réserve|reserve|^\d{2}\s\d{2})`)
	matchGroup         = regexp.MustCompile(`(?i)Gr\.*\s*(\d{1})`)
	matchGroupLocation = regexp.MustCompile(`(?i)Gr\.*\s*(\d{1})[:-|\s]*(salle|\s*)([A-Z]{1}\s*\d{3})`)
	matchTERLocation   = regexp.MustCompile(`(?i)TER\s*(\d{1})\n(salle)*\s*[:-]*\s*([A-Z]{1}\s*\d{3})`)
	matchTimeRange     = regexp.MustCompile(`(?i)(\d{1,2})\s?h\s?(\d{2})\s*[\-à]\s*(\d{1,2})\s?h\s?(\d{2})`)
	matchLocation      = regexp.MustCompile(`(?i)\s*(((Salle)*\s*:*\s*([A-Z]{1}\ *\d{3}))|(\w*[\s|\-]*amphi))`)
	cleanup            = regexp.MustCompile(`(?i)(option\s*:*\s\n*)`)
)

type Coords struct {
	X  int
	Y  int
	_Z int
}

type TimeRange struct {
	start time.Duration
	end   time.Duration
}

type Schedule struct {
	tables    []*Table
	daysIndex int
	lineIndex int
}

func newSchedule() *Schedule {
	return &Schedule{
		daysIndex: 0,
		lineIndex: 0,
	}
}

func (s *Schedule) addTable() {
	t := &Table{
		timeRanges: make(map[int]TimeRange),
		days:       make(map[int]string),
		dates:      make(map[int]string),
		entries:    make(map[Coords]string),
	}

	s.tables = append(s.tables, t)
}

func (s *Schedule) table() *Table {
	if len(s.tables) == 0 {
		panic("FAILURE: trying to insert things before a table header")
	}

	return s.tables[len(s.tables)-1]
}

func (s *Schedule) parseLine(cols []string) {
	for k, v := range cols {
		sv := strings.TrimSpace(v)

		if len(sv) > 0 {
			if isTimeRange.MatchString(sv) {
				s.table().timeRanges[k] = toTimeRange(matchTimeRange.FindAllStringSubmatch(sv, -1)[0])

				if debug {
					log.Printf("(%d:%d) Found a time range: '%s'\n", k, s.lineIndex, sv)
				}
			} else if isDay.MatchString(sv) {
				// New header = new table, support for multiple time tables in one schedule file
				if s.daysIndex != s.lineIndex {
					s.addTable()
					s.daysIndex = s.lineIndex

					if debug {
						log.Printf("(%d:%d) Switching to a new table\n", k, s.lineIndex)
					}
				}

				s.table().days[k] = sv

				if debug {
					log.Printf("(%d:%d) Found a day: '%s'\n", k, s.lineIndex, sv)
				}
			} else if isDate.MatchString(sv) {
				s.table().dates[s.lineIndex] = sv

				if debug {
					log.Printf("(%d:%d) Found a date: '%s'\n", k, s.lineIndex, sv)
				}
			} else if k > 0 && !isGarbage.MatchString(sv) {
				if debug {
					log.Printf("(%d:%d) Found an entry: '%s'\n", k, s.lineIndex, one(sv))
				}

				if isWrongLine(sv) {
					if debug {
						log.Printf("Wrong line detected, reintegrating\n")
					}

					// Reintegrate groups/location in the entry of the previous line
					s.table().splitEntry(k, s.lineIndex-1, s.table().entries[Coords{k, s.lineIndex - 1, 0}]+"\n"+sv)
				} else {
					s.table().splitEntry(k, s.lineIndex, sv)
				}
			}
		}
	}

	s.lineIndex++
}

func (s *Schedule) outputCalendar() {
	vcal := ical.NewBasicVCalendar()
	vcal.X_WR_CALNAME = calendarName

	for n, t := range s.tables {
		if debug {
			log.Printf("Starting output of table %d\n", n)
		}

		for k, v := range t.entries {
			var ve ical.VEvent
			var tr TimeRange

			d := t.getDate(k.X, k.Y)

			if matchTimeRange.MatchString(v) {
				tr = toTimeRange(matchTimeRange.FindAllStringSubmatch(v, -1)[0])
			} else {
				tr = mapFindLowerIntervalTR(t.timeRanges, k.X)
			}

			ve.DTSTART = d.Add(tr.start)
			ve.DTEND = d.Add(tr.end)

			if matchLocation.MatchString(v) {
				ve.LOCATION = formatLocation(matchLocation.FindAllStringSubmatch(v, -1)[0][1])
			}

			ve.SUMMARY = cleanup.ReplaceAllString(v, "")
			ve.SUMMARY = strings.Split(ve.SUMMARY, "\n")[0]
			ve.SUMMARY = matchGroup.ReplaceAllString(ve.SUMMARY, "") + " " + formatGroup(v)
			ve.SUMMARY = matchTimeRange.ReplaceAllString(ve.SUMMARY, "")
			ve.SUMMARY = matchLocation.ReplaceAllString(ve.SUMMARY, "")
			ve.SUMMARY = strings.Split(ve.SUMMARY, "-")[0]
			ve.SUMMARY = strings.Split(ve.SUMMARY, "+")[0]
			ve.SUMMARY = strings.TrimSpace(ve.SUMMARY)

			ve.TZID = timeZone

			ve.UID = ve.DTSTART.Format(uidDTformat) + "-" + strings.Replace(ve.SUMMARY, " ", "", -1) + uidHost

			vcal.VComponent = append(vcal.VComponent, &ve)
		}
	}

	var b bytes.Buffer

	if err := vcal.Encode(&b); err != nil {
		panic(err)
	}

	if debug {
		log.Printf("Finished output\n")
	} else {
		fmt.Print(b.String())
	}
}
