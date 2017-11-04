package main

import (
	"fmt"
	"io/ioutil"
)

func read() ([]byte, error) {
	b, err := ioutil.ReadFile("a.g")
	if err != nil {
		return nil, err
	}
	return b, nil
}

func main() {
	fmt.Println(read())
}
