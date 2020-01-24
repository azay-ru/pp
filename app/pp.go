package pp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/soniah/gosnmp"
)

var Counters VendorsMap

func (d *Device) Request(oids []string) (answer []string, err error) {
	gosnmp.Default.Target = d.Host
	answer = make([]string, 0, len(Fields))

	if err := gosnmp.Default.Connect(); err != nil {
		return answer, fmt.Errorf("Can't connect to %v: %v\n", d.Host, err)

	}
	defer gosnmp.Default.Conn.Close()

	if result, err := gosnmp.Default.Get(oids); err != nil {
		log.Printf("Error SNMP request to %v\n", d.Host)
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
		for _, i := range Fields { //Vendors[d.VendorID]  {
			oids = append(oids, Vendors[Devices[n].VendorID][i])
		}

		// Request
		if resp, err := Devices[n].Request(oids); err != nil {
			return err
		} else {
			Devices[n].Counter = resp

			// Build counters object
			fm := make(FieldsMap, len(Fields))
			for x, j := range Fields {
				fm[j] = resp[x]
			}
			Counters[Devices[n].Host] = fm
		}
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
		var line string

		if Config.Header {
			for dev := range Counters {
				for j := range Counters[dev] {
					line = line + Delimiter + j
				}
				line = line + "\n"
				break
			}
		}

		for dev := range Counters {
			line = line + dev
			for j := range Counters[dev] {
				line = line + Delimiter + Counters[dev][j]
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
