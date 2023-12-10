package main

import (
	"errors"
	"flag"
	customerimporter "github.com/radoslaw.ratyna/teamwork"
	"log"
	"os"
)

func main() {
	outputPath := flag.String("o", "", "Output path for result of customerimporter program")
	inputPath := flag.String("f", "", "Path to source CSV data")

	flag.Parse()

	if _, err := os.Stat(*inputPath); errors.Is(err, os.ErrNotExist) {
		log.Fatal("Missing source CVS file!")
	}

	result, err := customerimporter.CountEmailDomains(*inputPath)
	if err != nil {
		log.Fatal(err)
	}

	if *outputPath != "" {
		customerimporter.SaveResultToFile(*outputPath, result)
		log.Println("Saved result to file " + *outputPath)
	} else {
		customerimporter.SaveResultToOutput(os.Stdout, result)
	}
}
