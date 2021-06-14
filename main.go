package main

import (
	"github.com/cretz/go-scrap"
)

// This example creates a screenshot at screenshot.png or a given filename.

func main() {

	// Setup scrap library
	if err := scrap.MakeDPIAware(); err != nil {
		panic(err)
	}
}
