package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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
	Sx := 1.0 // N/mm2
	var m ortho.Model
	m.Init(1800, 1800, "BASE")
	m.Add(100, "STIFF1", 600, true)
	m.Add(100, "STIFF2", 1200, true)
	ps, rs := m.Generate(100)
	ts := ortho.Select(ps)

	thks := []struct {
		name  string
		value uint64
	}{
		// upper case
		// uniq names
		// no space
		{"BASE", 12},
		{"STIFF1", 10},
		{"STIFF2", 10},
	}

	var out string

	out += fmt.Sprintf("*NODE, NSET=NALL\n")
	for i := range ps {
		x, y, z := ps[i][0], ps[i][1], ps[i][2]
		// TODO imperfection:
		// if z == 100 {
		// 	y += uint64(float64(x)/1800.0*20)
		// }
		out += fmt.Sprintf("%5d,%5d,%5d,%5d\n", i+1, x, y, z)
	}

	// TODO : S4 or S4R
	for _, t := range thks {
		out += fmt.Sprintf("*ELEMENT, TYPE=S4, ELSET=%s\n", t.name)
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

	out += fmt.Sprintf("*MATERIAL, NAME=STEEL\n")
	out += fmt.Sprintf("*ELASTIC\n")
	out += fmt.Sprintf("205000,0.3\n")

	for _, t := range thks {
		out += fmt.Sprintf("*SHELL SECTION,ELSET=%s, MATERIAL=STEEL\n", t.name)
		out += fmt.Sprintf("%5d\n", t.value)
	}

	out += fmt.Sprintf("*STEP\n")
	out += fmt.Sprintf("*BUCKLE\n2\n") // 2 buckling modes

	out += fmt.Sprintf("*CLOAD\n")

	// Load on left/right edge
	for _, s := range []struct {
		pts          []ortho.PointType
		byX, sortByX bool
		factor       float64
	}{
		{
			pts:     []ortho.PointType{ortho.Left, ortho.LeftBottom, ortho.LeftTop},
			byX:     true,
			sortByX: false,
			factor:  Sx,
		},
		{
			pts:     []ortho.PointType{ortho.Right, ortho.RightBottom, ortho.RightTop},
			byX:     true,
			sortByX: false,
			factor:  -Sx,
		},
	} {
		if math.Abs(s.factor) == 0.0 {
			continue
		}
		// store indexes
		ind := []int{}
		for i := range ts {
			for _, pt := range s.pts {
				if pt == ts[i] {
					ind = append(ind, i)
				}
			}
		}
		// sorting
		sort.SliceStable(ind, func(i, j int) bool {
			i = ind[i]
			j = ind[j]
			if s.sortByX {
				return ps[i][0] < ps[j][0]
			}
			return ps[i][1] < ps[j][1]
		})
		// load distribution
		for i := range ind {
			var L float64
			if i != 0 {
				if s.byX {
					L += float64(ps[ind[i]][1]-ps[ind[i-1]][1]) / 2.0
				} else {
					L += float64(ps[ind[i]][0]-ps[ind[i-1]][0]) / 2.0
				}
			}
			if i != len(ind)-1 {
				if s.byX {
					L += float64(ps[ind[i+1]][1]-ps[ind[i]][1]) / 2.0
				} else {
					L += float64(ps[ind[i+1]][0]-ps[ind[i]][0]) / 2.0
				}
			}
			force := L * float64(thks[0].value) * s.factor
			if s.byX {
				out += fmt.Sprintf("%5d,%5d, %+9.5e\n", ind[i]+1, 1, force)
			} else {
				out += fmt.Sprintf("%5d,%5d, %+9.5e\n", ind[i]+1, 2, force)
			}
		}
	}

	// print stresses
	for _, t := range thks {
		out += fmt.Sprintf("*EL FILE  , ELSET=%s\nS\n", t.name)
		out += fmt.Sprintf("*EL PRINT , ELSET=%s\nS\n", t.name)
	}
	// print displacement
	out += fmt.Sprintf("*NODE FILE  , NSET=NALL\nU\n")
	out += fmt.Sprintf("*NODE PRINT , NSET=NALL\nU\n")
	out += fmt.Sprintf("*END STEP\n")

	output, err := ccx(out)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println(output)

	parse(output)
}

func ccx(content string) (output string, err error) {
	dir, err := ioutil.TempDir("", "ccx")
	if err != nil {
		return
	}

	defer os.RemoveAll(dir) // clean up

	tmpfn := filepath.Join(dir, "1")
	if err = ioutil.WriteFile(tmpfn+".inp", []byte(content), 0666); err != nil {
		return
	}

	_, err = exec.Command("ccx", "-i", tmpfn).Output()
	if err != nil {
		return
	}

	tmpdat := filepath.Join(dir, "1.dat")
	dat, err := ioutil.ReadFile(tmpdat)
	if err != nil {
		return
	}

	return string(dat), nil
}

// type Buckle struct {
// 	Factor        float64
// 	Displacements []Node
// }

func parse(content string) {
	content = strings.ReplaceAll(content, "\r", "")
	lines := strings.Split(content, "\n")

	//      B U C K L I N G   F A C T O R   O U T P U T
	//
	//  MODE NO       BUCKLING
	//                 FACTOR
	//
	//       1   0.2680174E+03
	//       2   0.3149906E+03
	//
	var buckle []float64
	for i := range lines {
		if !strings.Contains(lines[i], "B U C K L I N G   F A C T O R   O U T P U T") {
			continue
		}
		i += 5
		for ; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" {
				break
			}
			fields := strings.Fields(lines[i])
			if len(fields) != 2 {
				panic(lines[i])
			}
			factor, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				panic(fields[1])
			}
			buckle = append(buckle, factor)
		}
	}
	sort.Float64s(buckle) // sorting
	fmt.Println("Buckling factors: ", buckle[0])

	//  stresses (elem, integ.pnt.,sxx,syy,szz,sxy,sxz,syz) for set BASE and time  0.0000000E+00
	//
	//         33   1 -9.995126E-01 -1.893993E-03 -1.058313E-03 -5.286196E-02  3.408146E-04 -3.354961E-03
	//         33   2 -9.956041E-01 -1.324090E-04  3.801916E-04 -5.255651E-02 -5.595679E-03 -6.205827E-03
	type stress struct {
		name  string
		elem  int64
		pnt   int64
		value [6]float64
	}
	var stresses []stress
	var first string // name of first plates
	for i := range lines {
		prefix := "stresses (elem, integ.pnt.,sxx,syy,szz,sxy,sxz,syz) for set "
		if !strings.Contains(lines[i], prefix) {
			continue
		}
		name := strings.TrimSpace(lines[i][len(prefix):])
		index := strings.Index(name, " ")
		name = name[:index]
		if first == name {
			// search stresses only for first dataset
			break
		}
		if first == "" {
			first = name
		}
		i += 2
		for ; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" {
				break
			}
			fields := strings.Fields(lines[i])
			if len(fields) != 8 {
				panic(lines[i])
			}
			elem, err := strconv.ParseInt(fields[0], 10, 64)
			if err != nil {
				panic(fields[0])
			}
			pnt, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				panic(fields[1])
			}
			var value [6]float64
			for i := 2; i < 8; i++ {
				factor, err := strconv.ParseFloat(fields[i], 64)
				if err != nil {
					panic(fields[i])
				}
				value[i-2] = factor
			}
			stresses = append(stresses, stress{name: name, elem: elem, pnt: pnt, value: value})
		}
	}
	// fmt.Println(stresses)
	// calculate maximal stresses
	var Smax float64
	for i := range stresses {
		var s float64
		for _, v := range stresses[i].value {
			s += v * v
		}
		s = math.Sqrt(s)
		Smax = math.Max(Smax, s)
	}
	fmt.Println("Smax = ", Smax, "MPA")

	//  displacements (vx,vy,vz) for set NALL and time  0.0000000E+00
	//
	//          1 -2.883217E-04  8.952994E-04  0.000000E+00
	//          2 -5.853373E-03  8.084879E-04  1.284318E-05
}
