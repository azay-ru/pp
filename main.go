package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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
var Vendors VendorsMap
var Counters VendorsMap
var Formats []string = []string{"json", "csv"}

type VendorsMap map[string]FieldsMap
type FieldsMap map[string]string

/*
Key for VendorsMaps = vendor ID
  Value = slice of FieldsMap,
  where FieldsMap is map too, where Key = field name, value = OID for this field

example:
VendorsMap {
    "hp": {
        "model":"1.3.6.1.2.1.25.3.2.1.3.1",
		"cartridge":"1.3.6.1.2.1.43.11.1.1.6.1.1"
	}
}
*/

type ConfigType struct {
	SkipEmpty  bool
	DontDetect bool
	Output     string
	Format     int
	Header     bool
	// Safe       bool 	// to-do, maybe later
}

type Device struct {
	Host     string
	VendorID string
}

func (v *VendorsMap) Init(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Can't read Vendors file \"%s\"\n", file)
	}
	*v = make(VendorsMap)
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("Can't parsing Vendors file \"%s\": %v\n", file, err)
	}
	return nil
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
	var timeout int
	var help bool
	flag.BoolVar(&help, "h", false, "Show this help")
	flag.StringVar(&pDevices, "p", "", "list of print devices (IP address of network name), comma separated: <host1[:vendor]>,<host2[:vendor]>...")
	flag.StringVar(&fDevices, "f", "", "file that contain names or IP addresses of print devices, one per line <host>[:vendor]")
	flag.StringVar(&vFile, "v", "vendors.json", "Vendors file")
	flag.StringVar(&format, "format", "json", "Output format")
	flag.StringVar(&Config.Output, "o", "", "output file, default stdout")
	flag.StringVar(&fields, "fields", "", "set of fields, specified in Vendors file")
	flag.IntVar(&timeout, "t", 10, "Timeout in seconds")
	flag.BoolVar(&Config.Header, "header", false, "Insert header for CSV format")
	//flag.BoolVar(&Config.Safe, "safe", false, "Safe mode")	// to-do, maybe later
	flag.Parse()

	if help {
		flag.Usage()
		return nil
	}

	// Config output format
	for n, v := range Formats {
		if format == v {
			Config.Format = n
		}
	}

	if timeout < 1 && timeout > 60 {
		return fmt.Errorf("Incorrect timeout %v:\n", timeout)
	}
	gosnmp.Default.Timeout = time.Duration(timeout) * time.Second

	// Config from Vendors file (json)
	if err := Vendors.Init(vFile); err != nil {
		return err
	}

	// Config fields
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
	return nil
}

func GetFields(f string) error {
	for _, col := range strings.Split(f, ",") {
		if len(col) > 0 {
			Fields = append(Fields, col)
		}
	}
	if len(Fields) == 0 {
		return fmt.Errorf("Fields not defined. See Vendors file and set options -fields, full help with options -h ")
	}

	return nil
}

func (d *Device) Request(oids []string) (answer []string, err error) {
	gosnmp.Default.Target = d.Host
	answer = make([]string, 0, len(Fields))

	if err := gosnmp.Default.Connect(); err != nil {
		return answer, fmt.Errorf("Can't connect to %s\n", d.Host)
	}
	defer gosnmp.Default.Conn.Close()

	if result, err := gosnmp.Default.Get(oids); err != nil {
		//return answer, fmt.Errorf("Error SNMP request to %v\n", d.Host)
		return answer, err
	} else {
		for _, v := range result.Variables {
			answer = append(answer, DecodeASN1(v))
		}
	}
	return
}

func Count() error {
	for n := 0; n < len(Devices); n++ {

		// slice for all oids per one device
		oids := make([]string, 0, len(Fields))

		// oids collect all OIDs each
		for _, field := range Fields { //Vendors[d.VendorID]  {
			if oid, ok := Vendors[Devices[n].VendorID][field]; ok {
				oids = append(oids, oid)
			} else {
				return fmt.Errorf("Invalid field \"%s\"\n", field)
			}

		}

		// SNMP Requests
		var resp []string
		var err error
		if resp, err = Devices[n].Request(oids); err != nil {
			log.Output(0, err.Error())
			continue // Skip if error on request
		}

		if len(resp) != len(Fields) {
			log.Output(0, fmt.Sprintf("Error in response for %s\n", Devices[n].Host))
			continue // Skip if error in response
		}
		// Build counters object
		fm := make(FieldsMap, len(Fields))
		for x, j := range Fields {
			fm[j] = resp[x]
		}
		Counters[Devices[n].Host] = fm
	}
	return nil
}

func Export() error {
	var content []byte
	var err error

	switch Config.Format {
	// json
	case 0:
		content, err = json.MarshalIndent(Counters, "", "  ")
	// csv
	case 1:
		// For CSV output in the same order as input
		var line string
		if Config.Header {
			for _, f := range Fields {
				line = line + ";" + f
			}
			line = line + "\n"
		}

		for _, dev := range Devices {
			line = line + dev.Host
			for _, f := range Fields {
				line = line + Delimiter + Counters[dev.Host][f]
			}
			line = line + "\n"
		}
		content, err = []byte(line), nil
	}

	if err != nil {
		return err
	}

	var out *os.File
	if len(Config.Output) > 0 { // Save to file
		var err error
		out, err = os.Create(Config.Output)
		if err != nil {
			return fmt.Errorf("Error write to %s, %s\n", Config.Output, err)
		}
		defer out.Close()
	} else { // Save to stdout
		out = os.Stdout
	}

	if _, err := io.WriteString(out, string(content)); err != nil {
		return err
	}

	return nil
}

// DecodeASN1 convert ANS.1 field to printable type
func DecodeASN1(v gosnmp.SnmpPDU) string {
	// fmt.Printf("%# v\n", pretty.Formatter(v))	// debug print
	switch v.Type {
	case gosnmp.OctetString:
		return string(v.Value.([]byte))
	case gosnmp.Counter32:
		return strconv.FormatUint(uint64((v.Value.(uint))), 10)
	default:
		return ""
	}
}

func main() {
	// All error to stderr
	log.SetOutput(os.Stderr)

	// Init devices list, vendors
	if err := Config.Init(); err != nil {
		log.Fatalln(err)
	}

	if err := Count(); err != nil {
		log.Fatalln(err)
	}

	if err := Export(); err != nil {
		log.Fatalln(err)
	}
}
