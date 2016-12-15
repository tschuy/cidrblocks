package cloudformation

import (
	"bytes"
	"html/template"
	"net"

	"github.com/tschuy/cidrblocks/subnet"
)

type block struct {
	Addr net.IPNet
	Type string
}

func Output(sn subnet.Subnet) (string, error) {
	var buf bytes.Buffer
	tmplPreamble, err := template.New("preamble").Parse(`{
	"AWSTemplateFormatVersion" : "2010-09-09",
	"Resources" : {
		"newvpc" : {
			"Type" : "AWS::EC2::VPC",
			"Properties" : {
				"CidrBlock" : "{{.cidrblock}}"
			}
		}`)

	if err != nil {
		return "", err
	}

	tmplOutro, err := template.New("outro").Parse(`
	}
}
`)

	if err != nil {
		return "", err
	}

	tmplAZ, err := template.New("az").Parse(`,
		"az{{.az}}{{.function}}" : {
			"Type" : "AWS::EC2::Subnet",
			"Properties" : {
				"VpcId" : { "Ref" : "newvpc" },
				"CidrBlock" : "{{.cidrblockInner}}",
				"AvailabilityZone" : "{{.region}}{{.az}}"
			}
		}`)

	if err != nil {
		return "", err
	}

	tmplPreamble.Execute(&buf, map[string]string{"cidrblock": sn.VPC.String()})
	for k, v := range sn.AvailabilityZones {
		for _, t := range []block{{v.Public, "public"}, {v.Private, "private"}, {v.Protected, "protected"}} {
			tmplAZ.Execute(&buf, map[string]string{
				"region":         "us-east-1", // sane way to set this needed (or way to set AZ without region?)
				"az":             subnet.AZName(k),
				"cidrblockInner": t.Addr.String(),
				"function":       t.Type,
			})
		}
	}
	tmplOutro.Execute(&buf, nil)

	return buf.String(), nil
}
