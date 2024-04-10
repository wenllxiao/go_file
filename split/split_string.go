package main

import (
	"fmt"
	"strings"
)

func SplitString(str string, sep string) {
	data := strings.Split(str, sep)
	var x string
	for _, s := range data {
		x += fmt.Sprintf("'%s',", s)
	}
	fmt.Println(x)
}
