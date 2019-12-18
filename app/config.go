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

// For production use build scripts from /build directory
var version = "Undefined"

var Config ConfigType

type Device struct {
	Host     string
	VendorID string
}

type ConfigType struct {
	SkipEmpty  bool
	DontDetect bool
	Output     string
	Devices    []Device
}

func (c *ConfigType) Init() error {
	flag.Usage = func() {
		fmt.Printf("Printed pages, v.%s\nUsage:\npp -p <...> | -f <file> [-no][-skip][-v] [-vendors <file>] [-format xml|json|csv] [-o <file>]\n\n", version)
		flag.PrintDefaults()
	}

	var printers []string // temp buffer, collect data from -p and -f options
	var pPrinters string
	var fPrinters string
	var vendorsFile string
	var help bool
	flag.BoolVar(&help, "h", false, "Show this help")
	flag.StringVar(&pPrinters, "p", "", "list of print devices (IP address of network name), comma separated: <host1[:vendor]>,<host2[:vendor]>...")
	flag.StringVar(&fPrinters, "f", "", "file that contain names or IP addresses of print devices, one per line <host>[:vendor]")
	flag.StringVar(&vendorsFile, "v", "vendors.xml", "file with Vendors description")
	flag.StringVar(&Config.Output, "o", "", "output file, use \"-o now\" for current time filename. ")
	flag.Parse()

	if help {
		flag.Usage()
		return nil
	}

	// Add printers from -p <...>
	if len(pPrinters) > 0 {
		printers = strings.Split(pPrinters, ",")
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
		return fmt.Errorf("Nothing to do")
	}

	// Set Vendors from dirty data -p and -f options
	c.Devices = make([]Device, 0, 32)
	for _, row := range printers {
		col := strings.Split(row, ":")

		device := Device{}
		if len(col) >= 1 { // Take first param
			device.Host = col[0]

			if len(col) >= 2 { // Take second param
				device.VendorID = strings.ToLower(col[1])
			}
		}
		if len(device.Host) > 0 {
			c.Devices = append(c.Devices, device)
		}
	}

	// This params maybe better set from flags
	gosnmp.Default.Retries = 1
	gosnmp.Default.Timeout = 5 * time.Second
	gosnmp.Default.ExponentialTimeout = false

	// Read data from Vendors file
	if err := Vendor.Init(vendorsFile); err != nil {
		return err
	}

	// fmt.Printf("% #v\n", pretty.Formatter(v))

	return nil
}
