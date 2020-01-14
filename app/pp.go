package pp

import (
	"fmt"
	"log"
	"strconv"

	"github.com/soniah/gosnmp"
)

var Counters VendorsMap

// Request() send SNMP requests and Retrieve answer for one device
// func (d *Device) Request(oids []string) (answer []gosnmp.SnmpPDU, err error) {
func (d *Device) Request(oids []string) (answer []string, err error) {
	gosnmp.Default.Target = d.Host
	answer = make([]string, 0, len(Fields))

	if err := gosnmp.Default.Connect(); err != nil {
		return answer, fmt.Errorf("Can't connect to %v: %v\n", d.Host, err)

	}
	defer gosnmp.Default.Conn.Close()

	// var result gosnmp.SnmpPacket
	// var err error
	if result, err := gosnmp.Default.Get(oids); err != nil {
		log.Printf("Error SNMP request to %v\n", d.Host)
	} else {
		for _, v := range result.Variables {
			answer = append(answer, DecodeASN1(v))
		}
	}
	return
}

// GetCounters get counters from all defined devices
func Count() error {

	// for v := range(Vendors) {
	// 	for _, i := range Items {
	// 		fmt.Println(v, i, Vendors[v][i])
	// 	}
	// }

	Counters = make(VendorsMap)

	for _, d := range Devices {

		// slice for all oids per one device
		oids := make([]string, 0, len(Fields))
		// answer := make([]string, 0, len(Items))

		// oids collect all OIDs each
		for _, i := range Fields { //Vendors[d.VendorID]  {
			oids = append(oids, Vendors[d.VendorID][i])
			// fmt.Printf("%s=%s ", i, Vendors[d.VendorID][i])
		}

		// Request
		fmt.Println("Host", d.Host, d.VendorID)
		fmt.Println(oids, "\n")
		if resp, err := d.Request(oids); err != nil {
			return err
		} else {
			fmt.Println(resp)
		}

		// }

		// Safe mode = send one OID per one SNMP request
		if Config.Safe {

		} else {

		}

		// counter, err := d.Request()
		// if err != nil {
		// 	log.Println(err.Error())
		// }

		// Counters = append(Counters, counter)
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

// func ExportXML() error {

// 	type XML struct {
// 		XMLName  xml.Name  `xml:"devices"`
// 		Counters []Counter `xml:"device"`
// 	}
// 	export := XML{}
// 	var output []byte

// 	for _, c := range Counters {
// 		if c.ok {
// 			export.Counters = append(export.Counters, c)
// 		}
// 	}

// 	// fmt.Printf("%# v\n", pretty.Formatter(export))

// 	output, err := xml.MarshalIndent(export, "", "  ")
// 	if err != nil {
// 		log.Printf("error: %v\n", err)
// 	}

// 	if len(Config.Output) == 0 {
// 		os.Stdout.Write(output)
// 	} else if Config.Output == "now" {
// 		Config.Output = time.Now().Format("20060102-1504.xml")
// 		fmt.Println(Config.Output)
// 	}

// 	return nil
// }
