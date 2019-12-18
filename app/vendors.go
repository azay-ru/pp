package pp

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

var Vendor VendorType

type VendorType struct {
	XMLName xml.Name `xml:"vendors"`
	Text    string   `xml:",chardata"`
	Vendors []struct {
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

func (v *VendorType) Init(file string) error {
	xmlData, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Invalid options \"-v %s\"\n", file)
	}
	if err := xml.Unmarshal(xmlData, &v); err != nil {
		return err
	}
	return nil
}
