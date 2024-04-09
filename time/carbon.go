package main

import (
	"fmt"

	"github.com/golang-module/carbon"
)

func main() {
	dateNow := carbon.Now().ToShortDateString()
	fmt.Println(dateNow)
}
