package eplate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Konstantin8105/ortho"
)

type Design struct {
	W, H, Thk  uint64 // mm
	Stiffiners []Stiffiner
}

func (d Design) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "Base plate dimensions:\n")
	fmt.Fprintf(w, "|\tHeigth:\t%5d\tmm\t|\n", d.H)
	fmt.Fprintf(w, "|\tWidth:\t%5d\tmm\t|\n", d.W)
	fmt.Fprintf(w, "|\tThickness:\t%5d\tmm\t|\n", d.Thk)
	fmt.Fprintf(w, "\n")
	// Stiffiners
	if 0 < len(d.Stiffiners) {
		for i, st := range d.Stiffiners {
			fmt.Fprintf(w, "Stiffiner position: %d\n", i+1)
			fmt.Fprintf(w, "%s", st)
		}
	}
	w.Flush()
	return buf.String()
}

type Stiffiner struct {
	W, Thk       uint64 // mm
	Offset       uint64 // distance from zero point
	IsHorizontal bool
}

func (s Stiffiner) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "Stiffiner dimensions:\n")
	fmt.Fprintf(w, "|\tWidth:\t%5d\tmm\t|\n", s.W)
	fmt.Fprintf(w, "|\tThickness:\t%5d\tmm\t|\n", s.Thk)
	fmt.Fprintf(w, "|\tOffset from plane:\t%5d\tmm\t|\n", s.Offset)
	fmt.Fprintf(w, "|\tHorizontal:\t%v\t\t|\n", s.IsHorizontal)
	w.Flush()
	return buf.String()
}

type Load struct {
	Sx, Sy, Tau float64 // MPa
	Pressure    float64 // Lateral pressure, MPa
}

func (l Load) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "Loads:\n")
	fmt.Fprintf(w, "|\tSx:\t%8.3f\tMPa\t|\n", l.Sx)
	fmt.Fprintf(w, "|\tSy:\t%8.3f\tMPa\t|\n", l.Sy)
	fmt.Fprintf(w, "|\tTau:\t%8.3f\tMPa\t|\n", l.Tau)
	fmt.Fprintf(w, "|\tLateral pressure:\t%8.3f\tMPa\t|\n", l.Pressure)
	w.Flush()
	return buf.String()
}

type Config struct {
	FET         FiniteElementType
	MaxDistance uint64
	Elasticity  float64 // MPa
	Ratio       float64 // dimensionless
}

func DefaultConfig() (def *Config) {
	def = &Config{
		FET:         S4,
		MaxDistance: 0,
		Elasticity:  205000, // MPa
		Ratio:       0.3,
	}
	return
}

type FiniteElementType string

const (
	S4 FiniteElementType = "S4"
)

func Calculate(d Design, l Load, c *Config) (buckle, Smax, Dmax float64) {
	if c == nil {
		c = DefaultConfig()
	}
	// TODO : check input data

	var m ortho.Model
	name := func(i int) string {
		return fmt.Sprintf("STIFF%d", i)
	}
	basename := "BASE"
	m.Init(d.W, d.H, basename)
	for i := range d.Stiffiners {
		m.Add(
			d.Stiffiners[i].W,
			name(i),
			d.Stiffiners[i].Offset,
			d.Stiffiners[i].IsHorizontal,
		)
	}
	ps, rs := m.Generate(c.MaxDistance)
	ts := ortho.Select(ps)

	type elset struct {
		name string
		thk  uint64
	}
	var thks = []elset{elset{name: basename, thk: d.Thk}}
	for i := range d.Stiffiners {
		thks = append(thks, elset{
			name: name(i),
			thk:  d.Stiffiners[i].Thk,
		})
	}

	builder := strings.Builder{}
	add := func(format string, a ...interface{}) {
		builder.WriteString(fmt.Sprintf(format, a...))
	}

	add("*NODE, NSET=NALL\n")
	for i := range ps {
		x, y, z := ps[i][0], ps[i][1], ps[i][2]
		// TODO imperfection:
		add("%5d,%5d,%5d,%5d\n", i+1, x, y, z)
	}

	for _, t := range thks {
		add("*ELEMENT, TYPE=%s, ELSET=%s\n", string(c.FET), t.name)
		for i := range rs {
			if t.name != rs[i].Material {
				continue
			}
			add("%5d", i+1)
			for _, p := range rs[i].PointsId {
				add(",%5d", p+1)
			}
			add("\n")
		}
	}

	add("*BOUNDARY\n")
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
				add("%5d,%5d,%5d\n", i+1, p+1, p+1)
			}
		}
	}

	add("*MATERIAL, NAME=STEEL\n")
	add("*ELASTIC\n")
	add("%12.2f,%12.2f\n", c.Elasticity, c.Ratio)

	for _, t := range thks {
		add("*SHELL SECTION,ELSET=%s, MATERIAL=STEEL\n", t.name)
		add("%5d\n", t.thk)
	}

	add("*STEP\n")
	add("*BUCKLE\n2\n") // 2 buckling modes

	add("*CLOAD\n")

	// Load on left/right edge
	for _, s := range []struct {
		pts     []ortho.PointType
		loadByX bool
		factor  float64
	}{
		// Sx
		{
			pts: []ortho.PointType{
				ortho.Left, ortho.LeftBottom, ortho.LeftTop},
			loadByX: true,
			factor:  +l.Sx,
		},
		{
			pts: []ortho.PointType{
				ortho.Right, ortho.RightBottom, ortho.RightTop},
			loadByX: true,
			factor:  -l.Sx,
		},
		// Sy
		{
			pts: []ortho.PointType{
				ortho.Top, ortho.LeftTop, ortho.RightTop},
			loadByX: false,
			factor:  +l.Sy,
		},
		{
			pts: []ortho.PointType{
				ortho.Bottom, ortho.RightBottom, ortho.LeftBottom},
			loadByX: false,
			factor:  -l.Sy,
		},
		// Tau
		{
			pts: []ortho.PointType{
				ortho.Left, ortho.LeftBottom, ortho.LeftTop},
			loadByX: false,
			factor:  -l.Tau,
		},
		{
			pts: []ortho.PointType{
				ortho.Right, ortho.RightBottom, ortho.RightTop},
			loadByX: false,
			factor:  +l.Tau,
		},
		{
			pts: []ortho.PointType{
				ortho.Top, ortho.LeftTop, ortho.RightTop},
			loadByX: true,
			factor:  +l.Tau,
		},
		{
			pts: []ortho.PointType{
				ortho.Bottom, ortho.RightBottom, ortho.LeftBottom},
			loadByX: true,
			factor:  -l.Tau,
		},
		// TODO Lateral pressure
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
		var dx, dy uint64
		{
			xmin, xmax := ps[ind[0]][0], ps[ind[0]][0]
			ymin, ymax := ps[ind[0]][1], ps[ind[0]][1]
			for _, index := range ind {
				if v := ps[index][0]; v < xmin {
					xmin = v
				}
				if v := ps[index][0]; xmax < v {
					xmax = v
				}
				if v := ps[index][1]; v < ymin {
					ymin = v
				}
				if v := ps[index][1]; ymax < v {
					ymax = v
				}
			}
			dx = xmax - xmin
			dy = ymax - ymin
		}
		if (dx == 0 && dy == 0) || (dx != 0 && dy != 0) {
			panic(fmt.Errorf("dx=%d,dy=%d", dx, dy))
		}
		// sorting
		var sd int // direction of sorting
		if dx < dy {
			sd = 1
		}
		sort.SliceStable(ind, func(i, j int) bool {
			i = ind[i]
			j = ind[j]
			return ps[i][sd] < ps[j][sd]
		})
		// load distribution
		for i := range ind {
			var L float64
			if i != 0 {
				L += float64(ps[ind[i]][sd]-ps[ind[i-1]][sd]) / 2.0
			}
			if i != len(ind)-1 {
				L += float64(ps[ind[i+1]][sd]-ps[ind[i]][sd]) / 2.0
			}
			force := L * float64(thks[0].thk) * s.factor
			if s.loadByX {
				add("%5d,%5d, %+9.5e\n", ind[i]+1, 1, force)
			} else {
				add("%5d,%5d, %+9.5e\n", ind[i]+1, 2, force)
			}
		}
	}

	// print stresses
	for _, t := range thks {
		add("*EL FILE  , ELSET=%s\nS\n", t.name)
		add("*EL PRINT , ELSET=%s\nS\n", t.name)
	}
	// print displacement
	add("*NODE FILE  , NSET=NALL\nU\n")
	add("*NODE PRINT , NSET=NALL\nU\n")
	add("*END STEP\n")

	output, err := ccx(builder.String())
	if err != nil {
		fmt.Println(err)
	}
	return parse(output)
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

func parse(content string) (buckle, Smax, Dmax float64) {
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
	buckle = 1e10
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
			factor = math.Abs(factor)
			if factor < buckle {
				buckle = factor
			}
		}
	}

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
	// calculate maximal stresses
	for i := range stresses {
		var s float64
		for _, v := range stresses[i].value {
			s += v * v
		}
		s = math.Sqrt(s)
		Smax = math.Max(Smax, s)
	}

	//  displacements (vx,vy,vz) for set NALL and time  0.0000000E+00
	//
	//          1 -2.883217E-04  8.952994E-04  0.000000E+00
	//          2 -5.853373E-03  8.084879E-04  1.284318E-05
	type disp struct {
		node  int64
		value [3]float64
	}
	var disps []disp
	counter := 0
	for i := range lines {
		prefix := "displacements (vx,vy,vz) for set"
		if !strings.Contains(lines[i], prefix) {
			continue
		}
		if 0 < counter {
			// parse only first displacement
			break
		}
		i += 2
		for ; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" {
				break
			}
			fields := strings.Fields(lines[i])
			if len(fields) != 4 {
				panic(lines[i])
			}
			node, err := strconv.ParseInt(fields[0], 10, 64)
			if err != nil {
				panic(fields[0])
			}
			var value [3]float64
			for i := 1; i < 4; i++ {
				factor, err := strconv.ParseFloat(fields[i], 64)
				if err != nil {
					panic(fields[i])
				}
				value[i-1] = factor
			}
			disps = append(disps, disp{node: node, value: value})
		}
	}
	// calculate maximal displacement
	for i := range disps {
		var s float64
		for _, v := range disps[i].value {
			s += v * v
		}
		s = math.Sqrt(s)
		Dmax = math.Max(Dmax, s)
	}

	return
}
