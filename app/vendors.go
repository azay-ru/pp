package pp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type ItemsMap map[string]string
type VendorsMap map[string]ItemsMap

var Items []string
var Vendors VendorsMap

// var Vendor VendorType

// type VendorType struct {
// 	XMLName xml.Name `xml:"vendors"`
// 	Text    string   `xml:",chardata"`
// 	Vendors []struct {
// 		Text      string `xml:",chardata"`
// 		ID        string `xml:"id,attr"`
// 		Serial    string `xml:"serial"`
// 		Model     string `xml:"model"`
// 		Printed   string `xml:"printed"`
// 		Scanned   string `xml:"scanned,omitempty"`
// 		Cartridge struct {
// 			Text     string `xml:",chardata"`
// 			ID       string `xml:"id,omitempty"`
// 			Capacity string `xml:"capacity,omitempty"`
// 			Printed  string `xml:"printed,omitempty"`
// 			Color    string `xml:"color,omitempty"`
// 		} `xml:"cartridge,omitempty"`
// 	} `xml:"vendor"`
// }

func (v VendorsMap) Init(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Invalid options \"-v %s\"\n", file)
	}
	v = make(VendorsMap)
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("Error parsing vendors: %v\n", err)
	}

	for i := range v {
		fmt.Println(v[i]["printed"])
	}
	return nil
}
