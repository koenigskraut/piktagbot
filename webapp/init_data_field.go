package webapp

import (
	"bufio"
	"errors"
	"fmt"
	"sync"
)

type InitDataField interface {
	Name() string
	EncodeData(*bufio.Writer) error
	DecodeData(*bufio.Reader) error
}

func EncodeField(w *bufio.Writer, field InitDataField) error {
	_, err := fmt.Fprintf(w, "%s=", field.Name())
	if err != nil {
		return err
	}
	return field.EncodeData(w)
}

func DecodeField(r *bufio.Reader) (InitDataField, error) {
	name, err := readName(r)
	if err != nil {
		return nil, err
	}
	fieldFunc, ok := getMapTypes()[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("unknown field name: \"%s\"", name))
	}
	field := fieldFunc()
	if err := field.DecodeData(r); err != nil {
		return nil, err
	}
	return field, nil
}

var mapTypes map[string]func() InitDataField
var mapTypesOnce = sync.Once{}

func getMapTypes() map[string]func() InitDataField {
	mapTypesOnce.Do(func() {
		mapTypes = map[string]func() InitDataField{
			QueryIDName:  func() InitDataField { return &QueryID{} },
			AuthDateName: func() InitDataField { return &AuthDate{} },
			PrefixName:   func() InitDataField { return &Prefix{} },
			UserName:     func() InitDataField { return &User{} },
			HashName:     func() InitDataField { return &Hash{} },
		}
	})
	return mapTypes
}
