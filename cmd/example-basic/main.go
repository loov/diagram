package main

import (
	"io/ioutil"
	"log"

	"loov.dev/diagram"
	"loov.dev/diagram/sequence"
)

func main() {
	seq := sequence.New()
	seq.AutoSleep = 0.5
	seq.AutoDelay = 0.5
	seq.Add(
		sequence.Send("Client", "Server", "Init"),
		sequence.Send("Server", "DNS", "Update"),
		sequence.Send("DNS", "Server", "ACK"),
		sequence.Send("Server", "Client", "Data"),
		sequence.Send("Client", "Server", "Update").Sleeping(1).Delayed(3),
		sequence.Send("Client", "Server", "Update").Sleeping(-0.5).Delayed(1),
		sequence.Send("Client", "Server", "Update").Sleeping(-0.5).Delayed(2),
	)

	svg := diagram.NewSVG(seq.Size())
	seq.Draw(svg)

	err := ioutil.WriteFile("diagram.svg", svg.Bytes(), 0755)
	if err != nil {
		log.Fatal(err)
	}
}
