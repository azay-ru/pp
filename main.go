package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kr/pretty"
	snmp "github.com/soniah/gosnmp"
)

// PrintDevice Information for one print device
type PrintDevice struct {
	Host   string // IP or hostname (?!)
	vendor string
	Serial string
	Pages  int64
	ok     bool
}

var ErrUnknownVendor = errors.New("Unknown vendor")

var PrintDevices []PrintDevice

// var logerr *log.Logger
var version = "Undefined"
var Vendors = []string{"hp", "kyocera"}
var IPaddr, // Show IP address
	Serial, //  - serial number
	Name, //  - name of device
	HostName, //  - host name of device
	// Cartridge, //  - cartridge
	// Capacity, //  - cartridge Capacity
	// Color, //  - color of cartridge
	SkipEmpty,
	DontDetect,
	Verbose bool

func main() {
	fmt.Printf("Printed pages count, v.%s\n", version)

	err := getConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}

	for _, p := range PrintDevices {
		if err := GetData(&p); err != nil {
			log.Println(err.Error())
		} else {
			p.ok = true
			fmt.Printf("%v\n", p)
		}
	}

	fmt.Printf("===\n%# v\n", pretty.Formatter(PrintDevices))
}

// getConfig read params from command line and preset basic config
func getConfig() error {
	flag.Usage = func() {
		fmt.Println("Usage: ppc -p <...> | -f <file> [-n][-s][-v]")
		flag.PrintDefaults()
	}

	var printers []string
	var iPrinters,
		fPrinters string

	flag.BoolVar(&Verbose, "v", false, "Verbose mode")
	flag.BoolVar(&SkipEmpty, "s", false, "Skip unavailable or undetectable devices")
	flag.BoolVar(&DontDetect, "n", false, "Don't auto detect device vendor, if not specified")
	flag.StringVar(&iPrinters, "p", "", "IP address of printers, comma separated: <host1[:vendor]>,<host2[:vendor]>...")
	flag.StringVar(&fPrinters, "f", "", "file that contain name or IP addresses of print devices, one per line <host>[:vendor]")

	flag.Parse()

	// Add printers from -p <...>
	if len(iPrinters) > 0 {
		printers = strings.Split(iPrinters, ",")
		// if len(printers) == 0 {
		// 	return fmt.Errorf("Input list is empty")
		// }

	}
	// Add printers from -f <file>
	if len(fPrinters) > 0 {
		if _, err := os.Stat(fPrinters); os.IsExist(err) {
			return fmt.Errorf("File %v not found\n", fPrinters)
		}

		file, err := os.Open(fPrinters)
		if err != nil {
			return fmt.Errorf("Can't read file %v or file not exists\n", fPrinters)
		}
		defer file.Close()

		scan := bufio.NewScanner(file)
		scan.Split(bufio.ScanLines)
		for scan.Scan() {
			printers = append(printers, scan.Text())
		}
	}

	if len(printers) == 0 {
		flag.Usage()
		return fmt.Errorf("Printers list is empty")
	}

	// Set vendor for each printer
	for _, row := range printers {
		col := strings.Split(row, ":")

		var pd PrintDevice
		if len(col) >= 1 {
			pd.Host = col[0]

			if len(col) >= 2 {
				pd.vendor = strings.ToLower(col[1])
			}
		}
		if len(pd.Host) > 0 {
			PrintDevices = append(PrintDevices, pd)
		}
	}

	snmp.Default.Retries = 1
	snmp.Default.Timeout = 1 * time.Second
	snmp.Default.ExponentialTimeout = false
	return nil

}

// DiscoveryVendor try detect vendor type sending some SNMP request
func DiscoveryVendor() ([]string, error) {

	return []string{}, nil
}

// GetVendor set vendor type for device
func GetVendor(v string) ([]string, error) {
	var oids []string
	switch v {
	case "hp":
		oids = append(oids,
			"1.3.6.1.4.1.11.2.3.9.4.2.1.1.3.3.0", // Serial number
			"1.3.6.1.2.1.43.10.2.1.4.1.1")        // Printed pages
	case "kyocera":
		oids = append(oids,
			"1.3.6.1.2.1.43.5.1.1.17.1",
			"1.3.6.1.4.1.1347.42.2.1.1.1.6.1.1")
	default:
		if DontDetect {
			return oids, ErrUnknownVendor
		} else {
			return DiscoveryVendor()
		}
	}
	return oids, nil
}

// GetData request data from defined print devices
func GetData(pd *PrintDevice) error {
	// var oids []string
	oids, err := GetVendor(pd.vendor)
	if err != nil {
		return ErrUnknownVendor
	}

	// Use default SNMP object
	snmp.Default.Target = pd.Host
	if err := snmp.Default.Connect(); err != nil {
		return fmt.Errorf("Can't connect to %v: %v\n", pd.Host, err)
	}
	defer snmp.Default.Conn.Close()

	if result, err := snmp.Default.Get(oids); err != nil {
		return fmt.Errorf("Error SNMP request to %v\n", pd.Host)
	} else {
		for n, _ := range result.Variables {
			switch n {
			case 0:

				pd.Serial = "SN"
			case 1:
				pd.Pages = 1 //snmp.ToBigInt(v.Value)
			}
		}
	}
	return nil
}
