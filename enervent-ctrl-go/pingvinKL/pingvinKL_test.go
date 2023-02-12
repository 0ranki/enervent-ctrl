package pingvinKL

import (
	"fmt"
	"testing"
	"strconv"
)

func TestNewCoil(t *testing.T) {
	data := readCsvLines("../coils.csv")
	addr := data[1][0]
	symbol := data[1][0]
	description := data[1][0]

	coil := newCoil(addr, symbol, description)
	typ := fmt.Sprintf("%T", coil)
	// Assert newCoil returns pingvinKL.pingvinCoil
	if typ != "pingvinKL.pingvinCoil" {
		t.Errorf("newCoil returned %s, expecting pingvinCoil", typ)
	}
	// Assert Address is int and matches CSV
	addrtype := fmt.Sprintf("%T", coil.Address)
	if addrtype != "int" {
		t.Errorf("newCoil.Address is of type %s, expecting int", addrtype)
	}
	iaddr, _ := strconv.Atoi(addr)
	if coil.Address != iaddr {
		t.Errorf("coil.Address is %d, expecting %d", coil.Address, iaddr)
	}
	// Assert Symbol is string and matches CSV
	symboltype := fmt.Sprintf("%T", coil.Symbol)
	if symboltype!= "string" {
		t.Errorf("coil.Symbol is of type %s, expecting string", symboltype)
    }
	if coil.Symbol != symbol {
		t.Errorf("coil.Symbol is %s, expecting %s", coil.Symbol, symbol)
	}
	// Assert Description is string and matches CSV
	descriptiontype := fmt.Sprintf("%T", coil.Description)
	if descriptiontype!= "string" {
		t.Errorf("coil.Description is of type %s, expecting string", descriptiontype)
    }
	if coil.Description != description {
		t.Errorf("coil.Description is %s, expecting %s", coil.Description, description)
	}
	// Assert Value is boolean and false
	valuetype := fmt.Sprintf("%T", coil.Value)
	if valuetype!= "bool" {
		t.Errorf("coil.Value is of type %s, expecting bool", valuetype)
    }
	if coil.Value != false {
		t.Errorf("coil.Value is %t, expecting false", coil.Value)
	}
}
