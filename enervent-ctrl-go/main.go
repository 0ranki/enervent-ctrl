package main

import (
	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
)

func main() {
	pingvin := pingvinKL.New()
	// fmt.Println(pingvin.Coils)
	print(pingvin.Coils[1].Description)
}
