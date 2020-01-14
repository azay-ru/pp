package pp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var Vendors VendorsMap

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
