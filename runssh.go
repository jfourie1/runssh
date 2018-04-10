package main

import (
	"flag"
	"fmt"
)


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
}
