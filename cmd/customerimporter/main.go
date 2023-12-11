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
		log.Println("Missing source CVS file!")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	result, err := customerimporter.CountEmailDomains(*inputPath)
	if err != nil {
		log.Fatal(err)
	}

	if *outputPath != "" {
		err := customerimporter.SaveResultToFile(*outputPath, result)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Saved result to file " + *outputPath)
	} else {
		customerimporter.SaveResultToOutput(os.Stdout, result)
	}
}
