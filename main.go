package main

import (
	"encoding/csv"
	"flag"
	"io"
	"io/ioutil"
	"strings"
)

func main() {
	file := flag.String("file", "edt.csv", "The file to be processed")
	data, err := ioutil.ReadFile(*file)

	if err != nil {
		panic(err)
	}

	rawSchedule := csv.NewReader(strings.NewReader(string(data)))
	schedule := newSchedule()

	for {
		line, err := rawSchedule.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		schedule.parseLine(line)
	}

	schedule.outputCalendar()
}
