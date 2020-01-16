package main

import (
	"log"
	"os"

	pp "github.com/azay-ru/pp/app"
)

func main() {
	// All error to stderr
	log.SetOutput(os.Stderr)

	// Init devices list, vendors
	if err := pp.Config.Init(); err != nil {
		log.Fatalln(err)
	}

	if err := pp.Count(); err != nil {
		log.Fatalln(err)
	}

	if err := pp.Export(); err != nil {
		log.Fatalln(err)
	}
}
