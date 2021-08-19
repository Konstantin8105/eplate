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
	m.Add(100, "horizontal", 600, true)
	m.Add(100, "horizontal", 1200, true)
	ps, rs := m.Generate(500)
	mainPlate, left, right, top, bottom := ortho.Select(ps)

	var out string

	out += fmt.Sprintf("*NODE, nset=nall\n")
	for i := range ps {
		out += fmt.Sprintf("%5d,%5d,%5d,%5d\n", i+1, ps[i][0], ps[i][1], ps[i][2])
	}

	out += fmt.Sprintf("*ELEMENT, type=S4R, elset=eall\n")
	for i := range rs {
		out += fmt.Sprintf("%5d", i+1)
		for _, p := range rs[i].PointsId {
			out += fmt.Sprintf(",%5d", p+1)
		}
		out += fmt.Sprintf("\n")
	}

	out += fmt.Sprintf("*BOUNDARY\n")
	boundary := make([][6]bool, len(ps))
	for _, sl := range [][]int{left, right, top, bottom} {
		for _, ind := range sl {
			boundary[ind][2] = true // fix: Z
			boundary[ind][5] = true // fix: MZ
		}
	}
	for _, ind := range mainPlate {
		if ps[ind][0] == 0 && ps[ind][1] == 0 && ps[ind][2] == 0 {
			boundary[ind][0] = true // fix: X
			break
		}
	}
	{
		var r uint64 = 0
		for _, ind := range mainPlate {
			if v := ps[ind][0]; r < v {
				r = v
			}
		}
		for _, ind := range mainPlate {
			if (ps[ind][0] == r || ps[ind][0] == 0 ) && ps[ind][1] == 0 {
				boundary[ind][1] = true // fix: Y
			}
		}
	}
	for _, ind := range mainPlate {
		for p := 0; p < 6; p++ {
			if boundary[ind][p] == true {
				out += fmt.Sprintf("%5d,%5d,%5d\n", ind+1, p+1, p+1)
			}
		}
	}

	out += fmt.Sprintf("*MATERIAL, name=Steel\n")
	out += fmt.Sprintf("*ELASTIC\n")
	out += fmt.Sprintf("210000,0.3\n")
	out += fmt.Sprintf("*SHELL SECTION,ELSET=eall, material=steel\n")

	// TODO
	out += fmt.Sprintf("5\n")

	out += fmt.Sprintf("*STEP\n")

	out += fmt.Sprintf("*BUCKLE\n")

	out += fmt.Sprintf("2\n")

	out += fmt.Sprintf("*CLOAD\n")

	// TODO
	for _, ind := range left {
		out += fmt.Sprintf("%5d,1,720\n", ind+1,)
	}
	// out += fmt.Sprintf("6,3,-10000\n")

	out += fmt.Sprintf("*EL PRINT,elset=eall\n")

	out += fmt.Sprintf("s\n")

	out += fmt.Sprintf("*NODE FILE\n")

	out += fmt.Sprintf("u\n")

	out += fmt.Sprintf("*EL FILE, output=3D\n")

	out += fmt.Sprintf("s\n")

	out += fmt.Sprintf("*END STEP\n")

	fmt.Println(out)
}
