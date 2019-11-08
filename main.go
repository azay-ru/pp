package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	snmp "github.com/soniah/gosnmp"
)

// Exported format

const (
	Xml = iota
	Json
	CSV
)

var version = "0.0.1" // Use go build -ldflags "-X main.version=XX.YY.ZZ" to define your version

// PrintDevice Information for one print device
type PrintDevice struct {
	Host   string `xml:"host,attr" json:"host"`
	Serial string `xml:"serial" json:"id"`
	Pages  string `xml:"pages" json:"pages"`

	export byte   `xml:"-" json:"-"`
	vendor string `xml:"-" json:"-"`
	ok     bool   `xml:"-" json:"-"`
}

var PrintDevices []PrintDevice

var Vendors = []string{"hp", "kyocera"}
var IPaddr, // Show IP address
	Serial, //  - serial number
	Name, //  - name of device
	HostName, //  - host name of device
	// Cartridge, //  - cartridge
	// Capacity, //  - cartridge Capacity
	// Color, //  - color of cartridge
	SkipEmpty,
	DontDetect bool

var (
	ErrUnknownVendor = errors.New("Unknown vendor")
	JustHelp         = errors.New("Help")
)

func main() {
	log.SetOutput(os.Stderr)
	err := getConfig()
	if err == JustHelp {
		return
	} else if err != nil {
		log.Fatalf(err.Error())
	}

	// if len(PrintDevices) == 0 {
	// 	return
	// }

	for i := 0; i < len(PrintDevices); i++ {
		// for _, p := range PrintDevices {
		if err := GetData(&PrintDevices[i]); err != nil {
			log.Println(err.Error())
		} else {
			PrintDevices[i].ok = true
			// fmt.Printf("---\n%v\n---\n", PrintDevices[i])

		}
	}

	Export(&PrintDevices)
	// fmt.Printf("===\n%# v\n", pretty.Formatter(PrintDevices)) // debug print
}

// getConfig read params from command line and preset basic config
func getConfig() error {
	flag.Usage = func() {
		fmt.Printf("Printed pages, v.%s\nUsage:\npp -p <...> | -f <file> [-n][-s][-v] [-o <file>] [-format xml|json|csv]\n", version)
		flag.PrintDefaults()
	}

	var export, outfile string
	var printers []string
	var iPrinters,
		fPrinters string

	flag.BoolVar(&SkipEmpty, "s", false, "Skip unavailable or undetectable devices")
	flag.BoolVar(&DontDetect, "n", false, "Don't auto detect device vendor, if not specified")
	flag.StringVar(&iPrinters, "p", "", "IP address of printers, comma separated: <host1[:vendor]>,<host2[:vendor]>...")
	flag.StringVar(&fPrinters, "f", "", "file that contain names or IP addresses of print devices, one per line <host>[:vendor]")
	flag.StringVar(&export, "format", "xml", "output format: xml|json|csv")
	flag.StringVar(&outfile, "o", "", "output file name.")
	flag.Parse()

	// Add printers from -p <...>
	if len(iPrinters) > 0 {
		printers = strings.Split(iPrinters, ",")
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

	if len(printers) == 0 { //  && ((len(iPrinters) > 0) || (len(fPrinters) > 0)) {
		flag.Usage()
		return nil //fmt.Errorf("Devices list is empty")
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

// DecodeASN1 convert ANS.1 field to printable type
func DecodeASN1(v snmp.SnmpPDU) string {
	// fmt.Printf("%# v\n", pretty.Formatter(v))	// debug print
	switch v.Type {
	case snmp.OctetString:
		return string(v.Value.([]byte))
	case snmp.Counter32:
		return strconv.FormatUint(uint64((v.Value.(uint))), 10)
	default:
		return ""
	}
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
		for n, v := range result.Variables {
			switch n {
			case 0:
				pd.Serial = DecodeASN1(v)
			case 1:
				pd.Pages = DecodeASN1(v) //snmp.ToBigInt(v.Value)
			}
		}
	}
	return nil
}

func Export(pd *[]PrintDevice) error {
	type XML struct {
		XMLName     xml.Name      `xml:"devices"`
		PrintDevice []PrintDevice `xml:"device"`
	}
	export := XML{}
	var output []byte

	for _, v := range *pd {
		if v.ok {
			export.PrintDevice = append(export.PrintDevice, v)
		}
	}

	output, err := xml.MarshalIndent(export, "", "  ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	os.Stdout.Write(output)
	return nil
}
