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
	fmt.Fprintf(os.Stdout, "%s\n%s\n", d, l)
	buckle, Smax, Dmax := Calculate(d, l, nil)

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
	// | Sx:  1.000 MPa |
	// | Sy:  0.000 MPa |
	// | Tau: 0.000 MPa |
	//
	// | Buckling factor       269.93 times |
	// | Maximal stress          1.05 MPa   |
	// | Maximal deformation     0.03 mm    |
}
