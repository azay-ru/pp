package pp

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/soniah/gosnmp"
)

const Delimiter = ";"

// For production use build scripts from /build directory
var version = "Undefined"
var OIDs map[string]string
var Config ConfigType
var Fields []string
var Devices []Device
var Сounters VendorsMap
var Formats []string = []string{"json", "csv"}

type ConfigType struct {
	SkipEmpty  bool
	DontDetect bool
	Timeout    int
	Output     string
	Format     int
	Header     bool
	// Safe       bool 	// to-do, maybe later
}

type Device struct {
	Host     string
	VendorID string
	Counter  []string
}

func (c *ConfigType) Init() error {
	flag.Usage = func() {
		fmt.Printf("Printed pages, v.%s\nUsage:\npp -p <...> | -f <file> [-columns <...>] [-v <file>] [-t <n>] [-o <file>]\n\n", version)
		flag.PrintDefaults()
	}

	var printers []string // temp buffer, collect data from -p and -f options
	var pDevices string
	var fDevices string
	var vFile string
	var fields string
	var format string
	var help bool
	flag.BoolVar(&help, "h", false, "Show this help")
	flag.StringVar(&pDevices, "p", "", "list of print devices (IP address of network name), comma separated: <host1[:vendor]>,<host2[:vendor]>...")
	flag.StringVar(&fDevices, "f", "", "file that contain names or IP addresses of print devices, one per line <host>[:vendor]")
	flag.StringVar(&vFile, "v", "vendors.json", "Vendors file")
	flag.StringVar(&format, "format", "json", "Output format")
	flag.StringVar(&Config.Output, "o", "", "output file, default stdout")
	flag.StringVar(&fields, "fields", "", "set of fields, specified in Vendors file")
	flag.IntVar(&Config.Timeout, "t", 15, "Timeout in seconds")
	flag.BoolVar(&Config.Header, "header", false, "Insert header for CSV format")
	//flag.BoolVar(&Config.Safe, "safe", false, "Safe mode")	// to-do, maybe later
	flag.Parse()

	if help {
		flag.Usage()
		return nil
	}

	//
	for n, v := range Formats {
		if format == v {
			Config.Format = n
		}
	}

	if Config.Timeout < 0 && Config.Timeout > 60 {
		return fmt.Errorf("Incorrect timeout %v:\n", Config.Timeout)
	}

	if err := GetFields(fields); err != nil {
		return err
	}

	// Add devices from -p <...>
	if len(pDevices) > 0 {
		printers = strings.Split(pDevices, ",")
	}

	// Add devices from -f <file>
	if len(fDevices) > 0 {
		if _, err := os.Stat(fDevices); os.IsExist(err) {
			return fmt.Errorf("File %v not found:\n", fDevices)
		}

		file, err := os.Open(fDevices)
		if err != nil {
			return fmt.Errorf("Can't read file %v or file not exists:\n", fDevices)
		}
		defer file.Close()

		scan := bufio.NewScanner(file)
		scan.Split(bufio.ScanLines)
		for scan.Scan() {
			printers = append(printers, scan.Text())
		}
	}

	if len(printers) == 0 { //  && ((len(iPrinters) > 0) || (len(fPrinters) > 0)) {
		return fmt.Errorf("Device list is empty")
	}

	// Read data from Vendors file (json)
	if err := Vendors.Init(vFile); err != nil {
		return err
	}

	// Set Vendors from dirty data -p and -f options
	Devices = make([]Device, 0, 128)

	for _, row := range printers {

		// one valid dirty record is <host>:<vendor ID>
		col := strings.Split(row, ":")
		if len(col) >= 2 {

			// Host not empty and VendorID exists in Vendors Map
			if _, ok := Vendors[col[1]]; ok && len(col[0]) > 0 {
				device := Device{}
				device.Host = col[0]
				device.VendorID = col[1]
				Devices = append(Devices, device)
			}
		}
	}
	Counters = make(VendorsMap, len(Devices))

	// This params maybe better set from flags
	gosnmp.Default.Retries = 1
	gosnmp.Default.ExponentialTimeout = false

	// fmt.Printf("% #v\n", pretty.Formatter(v))
	gosnmp.Default.Timeout = time.Duration(Config.Timeout) * time.Second
	return nil
}

func GetFields(f string) error {
	for _, col := range strings.Split(f, ",") {
		if len(col) > 0 {
			Fields = append(Fields, col)
		}
	}
	if len(Fields) == 0 {
		return fmt.Errorf("Fields not defined. See Vendors file and set options -fields")
	}

	return nil
}
