package table

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"net"
	"strconv"

	"github.com/tschuy/cidrblocks/subnet"
)

func Output(sn subnet.Subnet, extras *[]net.IPNet) (string, error) {

	tmpl, err := template.New("table").Parse(`
AZ {{ .az }} ({{ .azblock }}):
	{{ .private}} (Private - {{ .privcount }} addresses)
	{{ .public }} (Public - {{ .pubcount }} addresses)
	{{ .protected }} (Protected - {{ .protcount }} addresses)
	{{ .extra }} (Extra - {{ .extracount }} addresses)
`)

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	vpcr := fmt.Sprintf("VPC Range - %s\n", sn.VPC.String())
	buf.WriteString(vpcr)

	for k, v := range sn.AvailabilityZones {

		infomap := make(map[string]string)
		infomap["az"] = subnet.AZName(k)

		infomap["azblock"] = v.AZBlock.String()

		infomap["private"] = v.Private.String()
		a, b := v.Private.Mask.Size()
		infomap["privcount"] = strconv.FormatFloat(math.Pow(2, float64(b-a)), 'f', 0, 64)

		infomap["public"] = v.Public.String()
		a, b = v.Public.Mask.Size()
		infomap["pubcount"] = strconv.FormatFloat(math.Pow(2, float64(b-a)), 'f', 0, 64)

		infomap["protected"] = v.Protected.String()
		a, b = v.Protected.Mask.Size()
		infomap["protcount"] = strconv.FormatFloat(math.Pow(2, float64(b-a)), 'f', 0, 64)

		infomap["extra"] = v.Extra.String()
		a, b = v.Extra.Mask.Size()
		infomap["extracount"] = strconv.FormatFloat(math.Pow(2, float64(b-a)), 'f', 0, 64)

		var tbl bytes.Buffer
		err = tmpl.Execute(&tbl, infomap)
		if err != nil {
			return "", err
		}

		buf.WriteString(tbl.String())
	}

	if len(*extras) != 0 {
		buf.WriteString("\nUnused blocks:\n\t")
		for _, e := range *extras {
			buf.WriteString(e.String())
			buf.WriteString("\n\t")
		}
	}

	return buf.String(), nil
}
