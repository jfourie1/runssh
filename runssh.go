package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"github.com/360EntSecGroup-Skylar/excelize"
)

type host struct {
	host,user,passwd string
	output []string
}

func readX(inputFile *string) ([]host, []string, error) {
	var hosts []host
	var commands []string

	xlsx,err := excelize.OpenFile(*inputFile)
	if err != nil {
		return hosts,commands,err
	}
	sheetName := xlsx.GetSheetName(1)
	if sheetName == "" {
		return hosts,commands,errors.New("Empty workbook")
	}
	rows := xlsx.GetRows(sheetName)
	if len(rows) < 1 {
		return hosts,commands,errors.New("Empty sheet")
	}
	if len(rows[0]) < 4 {
		return hosts,commands,errors.New("Need at least 1 command")
	}
	var ih, iu, ip int

	for _,c := range(rows[0][3:]) {
		commands = append(commands, c)
	}
	for i,c := range(rows[0][0:3]) {
		switch strings.ToLower(c) {
			case "host":
				ih = i
			case "username":
				iu = i
			case "password":
				ip = i
			case "user":
				iu = i
		}
	}
	for _,r := range(rows[1:]) {
		var h host
		h.host = r[ih]
		h.user = r[iu]
		h.passwd = r[ip]
		hosts = append(hosts, h)
	}
	return hosts,commands,nil
}

func main() {
	inputFile := flag.String("input", "", "Input file")
	outputFile := flag.String("output", "", "Output file, will overwrite input if not specified")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Input file not specified")
		flag.PrintDefaults()
	}
	if *outputFile == "" {
		*outputFile = *inputFile
	}
	hosts,commands,err := readX(inputFile)
	if err != nil {
		fmt.Println("Unable to read input file:",err)
		return
	}
	fmt.Print(commands)
	fmt.Print(hosts)
}
