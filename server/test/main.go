package main

import (
	"boyl/server/scan/metadata/steam"
	"encoding/json"
	"fmt"
)

func main() {
	p := steam.NewProvider()

	g, err := p.Find("Cyberpunk 2077", 2020)
	if err != nil {
		panic(err)
	}

	marshaled, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(marshaled))
}
