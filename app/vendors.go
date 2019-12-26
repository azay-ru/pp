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


func (v *VendorsMap) Init(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Invalid options \"-v %s\"\n", file)
	}
	*v = make(VendorsMap)
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("Error parsing vendors: %v\n", err)
	}
	return nil
}