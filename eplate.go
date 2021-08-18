package main

import (
	"fmt"
	"os"

	o "github.com/Konstantin8105/ortho"
)

func main() {
	var m ortho.Model
	fmt.Fprintf(os.Stdout, "Horizontal and Vertical\n")
	m.Init(1800, 1200, "base")
	m.Add(100, "horizontal", 600, true)
	m.Add(100, "vertical", 1000, false)
	view(m.Generate(0))
	debug(m.Generate(100))
	fmt.Fprintf(os.Stdout, "\n")
}
