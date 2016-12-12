package table

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"net"
	"strconv"
)

func Output(vpccidr *net.IPNet, alloc []map[string]*net.IPNet) (string, error) {

	var buf bytes.Buffer

	vpcr := fmt.Sprintf("VPC Range - %s\n", vpccidr.String())
	buf.WriteString(vpcr)

	for k, v := range alloc {
		tmpl, err := template.New("table").Parse(`
AZ {{ .az }} ({{ .azblock }}):
    {{ .private}} (Private - {{ .privcount }} addresses)
      {{ .public }} (Public - {{ .pubcount }} addresses)
        {{ .protected }} (Protected - {{ .protcount }} addresses)
`)
		if err != nil {
			panic(err)
		}
		infomap := make(map[string]string)
		infomap["az"] = string(k + 65)

		infomap["azblock"] = v["azblock"].String()

		infomap["private"] = v["private"].String()
		a, b := v["private"].Mask.Size()
		infomap["privcount"] = strconv.FormatFloat(math.Pow(2, float64(b-a)), 'f', 0, 64)

		infomap["public"] = v["public"].String()
		a, b = v["public"].Mask.Size()
		infomap["pubcount"] = strconv.FormatFloat(math.Pow(2, float64(b-a)), 'f', 0, 64)

		infomap["protected"] = v["protected"].String()
		a, b = v["protected"].Mask.Size()
		infomap["protcount"] = strconv.FormatFloat(math.Pow(2, float64(b-a)), 'f', 0, 64)

		var tbl bytes.Buffer
		err = tmpl.Execute(&tbl, infomap)
		if err != nil {
			return "", err
		}

		buf.WriteString(tbl.String())
	}
	return buf.String(), nil
}
