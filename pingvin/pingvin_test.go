package pingvin

import (
	"fmt"
	"strconv"
	"testing"
)

func TestNewCoil(t *testing.T) {
	data := readCsvLines("../coils.csv")
	addr := data[1][0]
	symbol := data[1][1]
	description := data[1][2]

	coil := newCoil(addr, symbol, description)
	typ := fmt.Sprintf("%T", coil)
	// Assert newCoil returns pingvinKL.pingvinCoil
	if typ != "pingvinKL.pingvinCoil" {
		t.Errorf("newCoil returned %s, expecting pingvinKL.pingvinCoil", typ)
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
	if symboltype != "string" {
		t.Errorf("coil.Symbol is of type %s, expecting string", symboltype)
	}
	if coil.Symbol != symbol {
		t.Errorf("coil.Symbol is %s, expecting %s", coil.Symbol, symbol)
	}

	// Assert Description is string and matches CSV
	descriptiontype := fmt.Sprintf("%T", coil.Description)
	if descriptiontype != "string" {
		t.Errorf("coil.Description is of type %s, expecting string", descriptiontype)
	}
	if coil.Description != description {
		t.Errorf("coil.Description is %s, expecting %s", coil.Description, description)
	}

	// Assert Value is boolean and false
	valuetype := fmt.Sprintf("%T", coil.Value)
	if valuetype != "bool" {
		t.Errorf("coil.Value is of type %s, expecting bool", valuetype)
	}
	if coil.Value != false {
		t.Errorf("coil.Value is %t, expecting false", coil.Value)
	}

	// Assert Reserved is bool and false
	reservedtype := fmt.Sprintf("%T", coil.Reserved)
	if reservedtype != "bool" {
		t.Errorf("coil.Reserved is of type %s, expecting bool", reservedtype)
	}
	if coil.Reserved != false {
		t.Errorf("coil.Reserved is %t, expecting false", coil.Reserved)
	}
}

func TestNewReservedCoil(t *testing.T) {
	data := readCsvLines("../coils.csv")
	addr := data[3][0]
	symbol := data[3][1]
	description := data[3][2]

	coil := newCoil(addr, symbol, description)
	// Assert Reserved is bool and true
	typ := fmt.Sprintf("%T", coil.Reserved)
	if typ != "bool" {
		t.Errorf("coil.Reserved is of type %s, expecting bool", typ)
	}
	if coil.Reserved != true {
		t.Errorf("coil.Reserved is %t, expecting true", coil.Reserved)
	}
}

func TestNewRegister(t *testing.T) {
	data := readCsvLines("../registers.csv")
	addr := data[4][0]
	symbol := data[4][1]
	regtype := data[4][2]
	description := data[4][6]

	hreg := newRegister(addr, symbol, regtype, description)

	// Assert newRegister returns pingvinKL.pingvinRegister
	typ := fmt.Sprintf("%T", hreg)
	if typ != "pingvinKL.pingvinRegister" {
		t.Errorf("newRegister returned %s, expecting pingvinKL.pingvinRegister", typ)
	}

	// Assert Address is int and matches CSV
	addrtype := fmt.Sprintf("%T", hreg.Address)
	if addrtype != "int" {
		t.Errorf("hreg.Address is of type %s, expecting int", addrtype)
	}
	iaddr, _ := strconv.Atoi(addr)
	if hreg.Address != iaddr {
		t.Errorf("hreg.Address is %d, expecting %d", hreg.Address, iaddr)
	}

	// Assert Symbol is string and matches CSV
	symboltype := fmt.Sprintf("%T", hreg.Symbol)
	if symboltype != "string" {
		t.Errorf("hreg.Symbol is of type %s, expecting string", symboltype)
	}
	if hreg.Symbol != symbol {
		t.Errorf("hreg.Symbol is %s, expecting %s", hreg.Symbol, symbol)
	}

	// Assert Description is string and matches CSV
	descriptiontype := fmt.Sprintf("%T", hreg.Description)
	if descriptiontype != "string" {
		t.Errorf("hreg.Description is of type %s, expecting string", descriptiontype)
	}
	if hreg.Description != description {
		t.Errorf("hreg.Description is %s, expecting %s", hreg.Description, description)
	}

	// Assert Value is int and 0
	valuetype := fmt.Sprintf("%T", hreg.Value)
	if valuetype != "int" {
		t.Errorf("hreg.Value is of type %s, expecting int", valuetype)
	}
	if hreg.Value != 0 {
		t.Errorf("hreg.Value is %d, expecting false", hreg.Value)
	}

	// Assert Reserved is bool and false
	reservedtype := fmt.Sprintf("%T", hreg.Reserved)
	if reservedtype != "bool" {
		t.Errorf("hreg.Reserved is of type %s, expecting bool", reservedtype)
	}
	if hreg.Reserved != false {
		t.Errorf("hreg.Reserved is %t, expecting false", hreg.Reserved)
	}

	// Assert Type is string and matches CSV
	hregtype := fmt.Sprintf("%T", hreg.Type)
	if hregtype != "string" {
		t.Errorf("hreg.Type is of type %s, expecting string", hregtype)
	}
	if hreg.Type != regtype {
		t.Errorf("hreg.Type is %s, expecting %s", hreg.Type, regtype)
	}
}
