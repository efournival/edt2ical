package main

import (
	"encoding/csv"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

var (
	file  string
	debug bool
)

func main() {
	flag.StringVar(&file, "file", "edt.csv", "The file to be processed")
	flag.BoolVar(&debug, "debug", false, "Whether to display debug info or not")
	flag.Parse()

	data, err := ioutil.ReadFile(file)

	if err != nil {
		panic(err)
	}

	rawSchedule := csv.NewReader(strings.NewReader(string(data)))
	schedule := newSchedule()

	if debug {
		log.Printf("Starting input\n")
	}

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
