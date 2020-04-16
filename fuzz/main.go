package main

import (
	"bufio"
	"crypto/rand"
	"io/ioutil"
	"log"

	"github.com/aaronjanse/3mux/ecma48"
)

func main() {
	log.SetOutput(ioutil.Discard)

	go fuzz()
	go fuzz()
	go fuzz()
	go fuzz()
	go fuzz()
	fuzz()
}

func fuzz() {
	random := bufio.NewReader(rand.Reader)
	out := make(chan ecma48.Output)

	p := ecma48.NewParser(false)
	go p.Parse(random, out)

	for range out {
	}
}
