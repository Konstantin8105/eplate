package eplate

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"
)

func Example() {
	d := Design{
		W:   1800,
		H:   1800,
		Thk: 12,
		Stiffiners: []Stiffiner{
			Stiffiner{
				W:            100,
				Thk:          10,
				Offset:       600,
				IsHorizontal: true,
			},
			Stiffiner{
				W:            100,
				Thk:          10,
				Offset:       1200,
				IsHorizontal: true,
			},
		},
	}
	l := Load{
		Sx:  1.0,
		Sy:  0.0,
		Tau: 0.0,
	}
	c := DefaultConfig()
	c.MaxDistance = 100
	fmt.Fprintf(os.Stdout, "%s\n%s\n", d, l)
	buckle, Smax, Dmax := Calculate(d, l, c)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "|\tBuckling factor\t%8.2f\ttimes\t|\n", buckle)
	fmt.Fprintf(w, "|\tMaximal stress\t%8.2f\tMPa\t|\n", Smax)
	fmt.Fprintf(w, "|\tMaximal deformation\t%8.2f\tmm\t|\n", Dmax)
	w.Flush()
	fmt.Fprintf(os.Stdout, "%s\n", buf.String())

	// Output:
	// Base plate dimensions:
	// | Heigth:     1800 mm |
	// | Width:      1800 mm |
	// | Thickness:    12 mm |
	//
	// Stiffiner position: 1
	// Stiffiner dimensions:
	// | Width:               100 mm |
	// | Thickness:            10 mm |
	// | Offset from plane:   600 mm |
	// | Horizontal:        true     |
	// Stiffiner position: 2
	// Stiffiner dimensions:
	// | Width:               100 mm |
	// | Thickness:            10 mm |
	// | Offset from plane:  1200 mm |
	// | Horizontal:        true     |
	//
	// Loads:
	// | Sx:                  1.000 MPa |
	// | Sy:                  0.000 MPa |
	// | Tau:                 0.000 MPa |
	// | Lateral pressure:    0.000 MPa |
	//
	// | Buckling factor       269.93 times |
	// | Maximal stress          1.05 MPa   |
	// | Maximal deformation     0.03 mm    |
}

func ExampleLateralPressure() {
	d := Design{
		W:   2000,
		H:   500,
		Thk: 5,
	}
	l := Load{
		Sx:       0.0,
		Sy:       0.0,
		Tau:      0.0,
		Pressure: 0.01,
	}
	c := DefaultConfig()
	c.MaxDistance = 50
	c.Elasticity = 206000
	c.Ratio = 0.3
	fmt.Fprintf(os.Stdout, "%s\n%s\n", d, l)
	buckle, Smax, Dmax := Calculate(d, l, c)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "|\tBuckling factor\t%8.2f\ttimes\t|\n", buckle)
	fmt.Fprintf(w, "|\tMaximal stress\t%8.2f\tMPa\t|\n", Smax)
	fmt.Fprintf(w, "|\tMaximal deformation\t%8.2f\tmm\t|\n", Dmax)
	w.Flush()
	fmt.Fprintf(os.Stdout, "%s\n", buf.String())

	//p := float64(l.Pressure)
	//b := float64(d.W)
	//a := float64(d.H)
	//t := float64(d.Thk)
	//E := float64(c.Elasticity)
	//fmt.Println(">> sigma = ", 0.75*p*a*a/(t*t*(1.61*math.Pow(a/b, 3)+1)))
	//fmt.Println(">> defle = ", 0.142*p*a*a*a*a/(E*t*t*t*(2.21*math.Pow(a/b, 3)+1)))
	//
	//betta := 0.7410
	//fmt.Println("S = ", betta*p*a*a/(t*t))
	//
	//alpha := 0.1400
	//fmt.Println("D = ", alpha*p*a*a*a*a/(E*t*t*t))

	// Output:
}
