package main

import (
	"fmt"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
)

func main() {
	pingvin := pingvinKL.New()
	pingvin.Update()
	for i := 0; i < len(pingvin.Coils); i++ {
		fmt.Println(pingvin.Coils[i].Symbol, pingvin.Coils[i].Value, pingvin.Coils[i].Description)
	}
}
