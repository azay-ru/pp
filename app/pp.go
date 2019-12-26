package pp

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/soniah/gosnmp"
)

// var  Cartridge, //  - cartridge
// 	// Capacity, //  - cartridge Capacity
// 	// Color, //  - color of cartridge
// 	DontDetect bool

// Exported format
const (
	XML = iota
	JSON
	CSV
)

var Counters []string	//[]Counter

// type Counter struct {
// 	Host         string `xml:"host,attr" json:"host"`
// 	Serial       string `xml:"serial" json:"id"`
// 	PrintedPages string `xml:"printed_pages" json:"printed_pages"`
// 	ScannedPages string `xml:"scanned_pages" json:"scanned_pages"`
// 	// CartridgePages string `xml:"cartridge_id" json:"cartridge_id"`

// 	export byte   `xml:"-" json:"-"`
// 	vendor string `xml:"-" json:"-"`
// 	ok     bool   `xml:"-" json:"-"`
// }

// Request() send SNMP requests and Retreive answer for one device
func (d *Device, oids *string) Request() (c Counter, err error) {
	gosnmp.Default.Target = d.Host

	if err := gosnmp.Default.Connect(); err != nil {
		return c, fmt.Errorf("Can't connect to %v: %v\n", d.Host, err)

	}
	defer gosnmp.Default.Conn.Close()

	// Vendor.Vendors[0].Cartridge.Printed
	if result, err := gosnmp.Default.Get([]string{"1.3.6.1.4.1.11.2.3.9.4.2.1.1.3.3.0", "1.3.6.1.2.1.43.10.2.1.4.1.1"}); err != nil {
		log.Printf("Error SNMP request to %v\n", d.Host)
	} else {
		for n, v := range result.Variables {
			switch n {
			case 0:
				c.Serial = DecodeASN1(v)
			case 1:
				c.PrintedPages = DecodeASN1(v) //snmp.ToBigInt(v.Value)
			}
		}
	}

	c.ok = true
	c.Host = d.Host
	return
}

// GetCounters get counters from all defined devices
func Count() error {

	// for v := range(Vendors) {
	// 	for _, i := range Items {
	// 		fmt.Println(v, i, Vendors[v][i])
	// 	}
	// }

	for _, d := range Config.Devices {
		fmt.Println()

		// Safe mode = send one OID per one SNMP request
		if Config.Safe {

		} else { 

		}

		counter, err := d.Request()
		if err != nil {
			log.Println(err.Error())
		}

		Counters = append(Counters, counter)
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

func ExportXML() error {

	type XML struct {
		XMLName  xml.Name  `xml:"devices"`
		Counters []Counter `xml:"device"`
	}
	export := XML{}
	var output []byte

	for _, c := range Counters {
		if c.ok {
			export.Counters = append(export.Counters, c)
		}
	}

	// fmt.Printf("%# v\n", pretty.Formatter(export))

	output, err := xml.MarshalIndent(export, "", "  ")
	if err != nil {
		log.Printf("error: %v\n", err)
	}

	if len(Config.Output) == 0 {
		os.Stdout.Write(output)
	} else if Config.Output == "now" {
		Config.Output = time.Now().Format("20060102-1504.xml")
		fmt.Println(Config.Output)
	}

	return nil
}
