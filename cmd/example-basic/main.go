package main

import (
	"io/ioutil"
	"log"

	"github.com/loov/diagram"
	"github.com/loov/diagram/sequence"
)

func main() {
	seq := sequence.New()
	seq.Add(
		sequence.FromTo("Client", "Server", "Init"),
		sequence.FromTo("Server", "Client", "Data"),
		sequence.FromTo("Client", "Server", "Update").WithDelay(3),
		sequence.FromTo("Client", "Server", "Update").WithDelay(2),
		sequence.FromTo("Client", "Server", "Update").WithDelay(1),
	)

	svg := diagram.NewSVG(seq.Size())
	seq.Draw(svg)

	err := ioutil.WriteFile("diagram.svg", svg.Bytes(), 0755)
	if err != nil {
		log.Fatal(err)
	}
}
