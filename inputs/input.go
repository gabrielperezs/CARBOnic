package inputs

import (
	"errors"
	"fmt"

	"github.com/gabrielperezs/CARBOnic/inputs/sqs"
	"github.com/gabrielperezs/CARBOnic/lib"
)

func Get(cfg interface{}) (lib.Input, error) {
	c := cfg.(map[string]interface{})
	if _, ok := c["Type"]; !ok {
		return nil, errors.New("Type not defined")
	}

	switch c["Type"] {
	case "sqs":
		return sqs.NewOrGet(c)
	default:
		return nil, fmt.Errorf("Plugin don't exists: %s", c["Type"])
	}
}
