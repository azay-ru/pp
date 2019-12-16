package pp

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

type Vendor struct {
	XMLName xml.Name `xml:"vendors"`
	Text    string   `xml:",chardata"`
	Vendor  []struct {
		Text      string `xml:",chardata"`
		ID        string `xml:"id,attr"`
		Serial    string `xml:"serial"`
		Printed   string `xml:"printed"`
		Scanned   string `xml:"scanned"`
		Cartridge struct {
			Text     string `xml:",chardata"`
			ID       string `xml:"id"`
			Capacity string `xml:"capacity"`
			Printed  string `xml:"printed"`
			Color    string `xml:"color"`
		} `xml:"cartridge"`
	} `xml:"vendor"`
}

var Vendors = Vendor{}

// ReadVendors reading data Vendors definition from xml file
func ReadVendors(file string) error {
	xmlContent, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Error reading vendors definition from %s\n", file)
	}

	return xml.Unmarshal(xmlContent, &Vendors)
}
