package cloudformation

import (
	"bytes"
	"html/template"
	"net"
	"strconv"

	"github.com/tschuy/cidrblocks/subnet"
)

type block struct {
	Addr net.IPNet
	Type string
}

const (
	PREAMBLE = `{
	"AWSTemplateFormatVersion" : "2010-09-09",
	"Resources" : {
		"vpc" : {
			"Type" : "AWS::EC2::VPC",
			"Properties" : {
				"CidrBlock" : "{{.cidrblock}}",
				"EnableDnsHostnames": true,
				"EnableDnsSupport": true
			}
		},
		"internetroute" : {
				"Type" : "AWS::EC2::Route",
				"DependsOn" : "internetgateway",
				"Properties" : {
						"RouteTableId" : { "Ref" : "internetroutetable" },
						"DestinationCidrBlock" : "0.0.0.0/0",
						"GatewayId" : { "Ref" : "internetgateway" }
				}
		},
		"internetgateway" : {
			"Type" : "AWS::EC2::InternetGateway"
		},
		"internetroutetable" : {
			"Type" : "AWS::EC2::RouteTable",
			"Properties" : {
					"VpcId" : { "Ref" : "vpc" }
			}
		},
		"AttachGateway" : {
			"Type" : "AWS::EC2::VPCGatewayAttachment",
			"Properties" : {
					"VpcId" : { "Ref" : "vpc" },
					"InternetGatewayId" : { "Ref" : "internetgateway" }
			}
		}`

	AZ = `,
		"natgateway{{.az}}" : {
			"DependsOn" : "AttachGateway",
			"Type" : "AWS::EC2::NatGateway",
			"Properties" : {
				"AllocationId" : { "Fn::GetAtt" : ["eipnat{{.az}}", "AllocationId"]},
				"SubnetId" : { "Ref" : "az{{.az}}public"}
			}
		},
		"eipnat{{.az}}" : {
			"Type" : "AWS::EC2::EIP",
			"Properties" : {
				"Domain" : "vpc"
			},
			"DependsOn" : "AttachGateway"
		},
		"route{{.az}}" : {
			"Type" : "AWS::EC2::Route",
			"Properties" : {
				"RouteTableId" : { "Ref" : "privateroutetable{{.az}}" },
				"DestinationCidrBlock" : "0.0.0.0/0",
				"NatGatewayId" : { "Ref" : "natgateway{{.az}}" }
			}
		},
		"privateroutetable{{.az}}" : {
			"Type" : "AWS::EC2::RouteTable",
			"Properties" : {
				"VpcId" : { "Ref" : "vpc" }
			}
		},
		"protectedroutetable{{.az}}" : {
			"Type" : "AWS::EC2::RouteTable",
			"Properties" : {
				"VpcId" : { "Ref" : "vpc" }
			}
		},
		"az{{.az}}protectedsubnetrouteassociation" : {
			"Type" : "AWS::EC2::SubnetRouteTableAssociation",
			"Properties" : {
				"SubnetId" : { "Ref" : "az{{.az}}protected" },
				"RouteTableId" : { "Ref" : "protectedroutetable{{.az}}" }
			}
		},
		"az{{.az}}privatesubnetrouteassociation" : {
			"Type" : "AWS::EC2::SubnetRouteTableAssociation",
			"Properties" : {
				"SubnetId" : { "Ref" : "az{{.az}}private" },
				"RouteTableId" : { "Ref" : "privateroutetable{{.az}}" }
			}
		},
		"az{{.az}}publicsubnetrouteassociation" : {
			"Type" : "AWS::EC2::SubnetRouteTableAssociation",
			"Properties" : {
				"SubnetId" : { "Ref" : "az{{.az}}public" },
				"RouteTableId" : { "Ref" : "internetroutetable" }
			}
		}`

	SUBNET = `,
		"az{{.az}}{{.function}}" : {
			"Type" : "AWS::EC2::Subnet",
			"Properties" : { {{if eq .function "public"}}
				"MapPublicIpOnLaunch": true,{{end}}
				"VpcId" : { "Ref" : "vpc" },
				"CidrBlock" : "{{.cidrblockInner}}",
				"AvailabilityZone" : {
					"Fn::Select" : [ "{{.az}}", { "Fn::GetAZs" : "" } ]
				}
			}
		}`

	OUTRO = `
	}
}`
)

func Output(sn subnet.Subnet, extras *[]net.IPNet) (string, error) {
	var buf bytes.Buffer

	tmplPreamble, err := template.New("preamble").Parse(PREAMBLE)
	if err != nil {
		return "", err
	}

	tmplAZ, err := template.New("az").Parse(AZ)
	if err != nil {
		return "", err
	}

	tmplOutro, err := template.New("outro").Parse(OUTRO)
	if err != nil {
		return "", err
	}

	tmplSubnet, err := template.New("subnet").Parse(SUBNET)
	if err != nil {
		return "", err
	}

	tmplSubnet = tmplSubnet.Funcs(template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	})

	tmplPreamble.Execute(&buf, map[string]string{"cidrblock": sn.VPC.String()})
	for k, v := range sn.AvailabilityZones {
		for _, t := range []block{{v.Public, "public"}, {v.Private, "private"}, {v.Protected, "protected"}} {
			tmplSubnet.Execute(&buf, map[string]string{
				"region":         "us-east-1", // sane way to set this needed (or way to set AZ without region?)
				"cidrblockInner": t.Addr.String(),
				"function":       t.Type,
				"az":             strconv.Itoa(k),
			})
		}
		tmplAZ.Execute(&buf, map[string]string{
			"az": strconv.Itoa(k),
		})
	}
	tmplOutro.Execute(&buf, nil)

	return buf.String(), nil
}
