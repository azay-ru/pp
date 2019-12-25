package main

import (
	"log"
	"os"

	pp "github.com/azay-ru/pp/app"
)

func main() {
	log.SetOutput(os.Stderr)

	if err := pp.Config.Init(); err != nil {
		log.Fatalln(err)
	}

	if err := pp.GetCounters(); err != nil {
		log.Fatalln(err)
	}

	// pp.ExportXML()

}
