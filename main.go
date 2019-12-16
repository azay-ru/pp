package main

import (
	"log"
	"os"

	pp "github.com/azay-ru/pp/app"
)

//

func main() {
	log.SetOutput(os.Stderr)

	if err := pp.SetOpts(); err != nil {
		log.Println(err)
		return
	}
	return

	// if err := pp.GetConfig(version); err != nil {
	// 	log.Fatalf(err.Error())
	// }

	// if len(PrintDevices) == 0 {
	// 	return
	// }

	// for i := 0; i < len(PrintDevices); i++ {
	// 	// for _, p := range PrintDevices {
	// 	if err := GetData(&PrintDevices[i]); err != nil {
	// 		log.Println(err.Error())
	// 	} else {
	// 		PrintDevices[i].ok = true
	// 		// fmt.Printf("---\n%v\n---\n", PrintDevices[i])
	// 	}
	// }

	// Export(&PrintDevices)
	// // fmt.Printf("===\n%# v\n", pretty.Formatter(PrintDevices)) // debug print
}
