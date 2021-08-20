package main

import (
	"fmt"

	"github.com/Konstantin8105/ortho"
)

// *NODE, nset=nall
// 1,0,0,0
// 2,200,0,0
// 3,200,50,0
// 4,0,50,0
// 5,400,0
// 6,400,50
// *ELEMENT, type=S4R, elset=eall
// 1,1,2,3,4
// 2,2,5,6,3
// *BOUNDARY
// 1,1,6
// 4,1,6
// *MATERIAL, name=Steel
// *ELASTIC
// 210000,0.3
// *SHELL SECTION,ELSET=eall, material=steel
// 5
// *STEP
// *BUCKLE
// 2
// *CLOAD
// 6,3,-10000
// *EL PRINT,elset=eall
// s
// *NODE FILE
// u
// *EL FILE, output=3D
// s
// *END STEP

func main() {
	var m ortho.Model
	m.Init(1800, 1800, "base")
	m.Add(100, "stiff", 600, true)
	m.Add(100, "stiff", 1200, true)
	ps, rs := m.Generate(50)
	ts := ortho.Select(ps)

	thks := []struct {
		name  string
		value uint64
	}{
		{"base", 12},
		{"stiff", 10},
	}

	var out string

	out += fmt.Sprintf("*NODE, nset=nall\n")
	for i := range ps {
		out += fmt.Sprintf("%5d,%5d,%5d,%5d\n", i+1, ps[i][0], ps[i][1], ps[i][2])
	}

	// S4 or S4R
	for _, t := range thks {
		out += fmt.Sprintf("*ELEMENT, type=S4, elset=%s\n", t.name)
		for i := range rs {
			if t.name != rs[i].Material {
				continue
			}
			out += fmt.Sprintf("%5d", i+1)
			for _, p := range rs[i].PointsId {
				out += fmt.Sprintf(",%5d", p+1)
			}
			out += fmt.Sprintf("\n")
		}
	}

	out += fmt.Sprintf("*BOUNDARY\n")
	boundary := make([][6]bool, len(ps))
	for i := range ts {
		if ts[i] == ortho.Other || ts[i] == ortho.MainPlate {
			// do nothing
			continue
		}
		boundary[i][2] = true // fix: Z
		boundary[i][5] = true // fix: MZ
		switch ts[i] {
		case ortho.Left:
		case ortho.Right:
		case ortho.Top:
		case ortho.Bottom:
		case ortho.LeftTop:
		case ortho.LeftBottom:
			boundary[i][0] = true // fix: X
			boundary[i][1] = true // fix: Y
		case ortho.RightTop:
		case ortho.RightBottom:
			boundary[i][1] = true // fix: Y
		default:
			panic(ts[i])
		}
	}
	for i := range boundary {
		for p := 0; p < 6; p++ {
			if boundary[i][p] == true {
				out += fmt.Sprintf("%5d,%5d,%5d\n", i+1, p+1, p+1)
			}
		}
	}

	out += fmt.Sprintf("*MATERIAL, name=Steel\n")
	out += fmt.Sprintf("*ELASTIC\n")
	out += fmt.Sprintf("205000,0.3\n")

	for _, t := range thks {
		out += fmt.Sprintf("*SHELL SECTION,ELSET=%s, material=steel\n", t.name)
		out += fmt.Sprintf("%5d\n", t.value)
	}

	out += fmt.Sprintf("*STEP\n")

	out += fmt.Sprintf("*BUCKLE\n")

	out += fmt.Sprintf("2\n")

	out += fmt.Sprintf("*CLOAD\n")
	leftamount := 0
	for i := range ts {
		if ts[i] == ortho.Left {
			leftamount++
		}
	}
	leftamount += 2
	sigma := 1.0 // N/mm2
	force := 1800.0 * 12.0 * sigma / float64(leftamount-1)

	// TODO
	for i := range ts {
		if ts[i] == ortho.Left {
			out += fmt.Sprintf("%5d,1,%+.6e\n", i+1, force)
		}
		if ts[i] == ortho.LeftBottom {
			out += fmt.Sprintf("%5d,1,%++.6e\n", i+1, force/2)
		}
		if ts[i] == ortho.LeftTop {
			out += fmt.Sprintf("%5d,1,%+.6e\n", i+1, force/2)
		}
		if ts[i] == ortho.Right {
			out += fmt.Sprintf("%5d,1,%+.6e\n", i+1, -force)
		}
		if ts[i] == ortho.RightTop {
			out += fmt.Sprintf("%5d,1,%+.6e\n", i+1, -force/2)
		}
		if ts[i] == ortho.RightBottom {
			out += fmt.Sprintf("%5d,1,%+.6e\n", i+1, -force/2)
		}
	}
	// out += fmt.Sprintf("6,3,-10000\n")

// 	out += fmt.Sprintf("*EL PRINT,elset=eall\n")
// 
// 	out += fmt.Sprintf("s\n")

	out += fmt.Sprintf("*NODE FILE\n")

	out += fmt.Sprintf("u\n")

 	out += fmt.Sprintf("*EL FILE, output=3D\n")
 
 	out += fmt.Sprintf("s\n")

	out += fmt.Sprintf("*END STEP\n")

	fmt.Println(out)
}
