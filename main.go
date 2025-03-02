package main

import (
	"fmt"
	"strings"
)

func fuck(arg ...interface{}) {
	a := arg[0].(string)
	fmt.Println(strings.Split(a, ","))

}

func main() {
	//cmd.Start()
	fuck("dad dd", 22, 's')

}
