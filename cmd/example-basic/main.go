package main

import (
	"io/ioutil"
	"log"

	"github.com/loov/diagram"
	"github.com/loov/diagram/sequence"
)

func main() {
	seq := sequence.New()
	seq.AutoSleep = 0.5
	seq.AutoDelay = 0.5
	seq.Add(
		sequence.FromTo("Client", "Server", "Init"),
		sequence.FromTo("Server", "DNS", "Update"),
		sequence.FromTo("DNS", "Server", "ACK"),
		sequence.FromTo("Server", "Client", "Data"),
		sequence.FromTo("Client", "Server", "Update").WithSleep(1).WithDelay(3),
		sequence.FromTo("Client", "Server", "Update").WithSleep(-0.5).WithDelay(1),
		sequence.FromTo("Client", "Server", "Update").WithSleep(-0.5).WithDelay(2),
	)

	svg := diagram.NewSVG(seq.Size())
	seq.Draw(svg)

	err := ioutil.WriteFile("diagram.svg", svg.Bytes(), 0755)
	if err != nil {
		log.Fatal(err)
	}
}
