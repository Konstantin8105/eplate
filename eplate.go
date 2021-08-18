package main

import (
	"fmt"

	"github.com/Konstantin8105/ortho"
)

func main() {
	var m ortho.Model
	m.Init(1800, 1800, "base")
	m.Add(100, "horizontal", 600, true)
	m.Add(100, "horizontal", 1200, true)
	p, r := m.Generate(250)
	fmt.Println(p)
	fmt.Println(r)
	fmt.Println(">>", len(p), len(r))
}
