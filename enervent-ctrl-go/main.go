package main

import (
	"fmt"
	"log"
	"time"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
)

func main() {
	log.Println(time.Now())
	pingvin := pingvinKL.New()
	log.Println(time.Now())
	fmt.Println(pingvin.ReadCoil(40))
	log.Println(time.Now())
}
