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

var version = "Undefined" // For production use build scripts from /build directory

var Opts struct {
	SkipEmpty  bool
	DontDetect bool
	Devices    []struct {
		Addr   string
		Vendor string
	}
}

func SetOpts() error {
	flag.Usage = func() {
		fmt.Printf("Printed pages, v.%s\nUsage:\npp -p <...> | -f <file> [-no][-skip][-v] [-vendors <file>] [-format xml|json|csv] [-o <file>]\n\n", version)
		flag.PrintDefaults()
	}

	var printers []string
	var pPrinters string
	var fPrinters string
	var help bool
	flag.BoolVar(&help, "h", false, "Show this help")
	flag.StringVar(&pPrinters, "p", "", "list of print devices (IP address of network name), comma separated: <host1[:vendor]>,<host2[:vendor]>...")
	flag.StringVar(&fPrinters, "f", "", "file that contain names or IP addresses of print devices, one per line <host>[:vendor]")
	flag.Parse()

	if help {
		flag.Usage()
		return nil
	}

	// Add printers from -p <...>
	if len(fPrinters) > 0 {
		printers = strings.Split(fPrinters, ",")
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

	gosnmp.Default.Retries = 1
	gosnmp.Default.Timeout = 1 * time.Second
	gosnmp.Default.ExponentialTimeout = false

	fmt.Println(Opts)

	return nil
}
