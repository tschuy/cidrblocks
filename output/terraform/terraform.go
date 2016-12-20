package terraform

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
	PREAMBLE = `variable "cidr_block" {
    type = "string"
    default = "{{.cidrblock}}"
}

# Specify the provider and access details
provider "aws" {

}

data "aws_region" "default" {
  current = true
}

# current availability zones
data "aws_availability_zones" "available" {}

# Create a VPC to launch our instances into
resource "aws_vpc" "default" {
    cidr_block = "${var.cidr_block}"
    enable_dns_hostnames = true
}

# Grant the VPC internet access on its main route table
resource "aws_route" "internet_access" {
    route_table_id         = "${aws_route_table.route_table_public.id}"
    destination_cidr_block = "0.0.0.0/0"
    gateway_id             = "${aws_internet_gateway.default.id}"
}

resource "aws_internet_gateway" "default" {
  vpc_id = "${aws_vpc.default.id}"

    tags {
            Name = "vpc-igw"
          }
}

resource "aws_route_table" "route_table_public" {
    vpc_id = "${aws_vpc.default.id}"
    route {
        cidr_block = "0.0.0.0/0"
        gateway_id = "${aws_internet_gateway.default.id}"
    }
}
`

	SUBNET = `

resource "aws_subnet" "az_{{.az}}_{{.function}}" {
    vpc_id                  = "${aws_vpc.default.id}"
    cidr_block              = "{{.cidrblockInner}}"
    availability_zone       = "${data.aws_availability_zones.available.names[{{.az}}]}"
    map_public_ip_on_launch = {{if eq .function "public"}}true{{else}}false{{end}}
}

resource "aws_route_table_association" "association_{{.az}}_{{.function}}" {
    subnet_id = "${aws_subnet.az_{{.az}}_{{.function}}.id}"
	  depends_on = ["aws_route_table.route_table_{{.function}}{{if eq .function "public"}}{{else}}_{{.az}}{{end}}"]
    route_table_id = "${aws_route_table.route_table_{{.function}}{{if eq .function "public"}}{{else}}_{{.az}}{{end}}.id}"
}
`
	ROUTING = `

resource "aws_nat_gateway" "nat_gateway_{{.az}}" {
  allocation_id = "${aws_eip.eip_nat_{{.az}}.id}"
  subnet_id = "${aws_subnet.az_{{.az}}_public.id}"

  depends_on = ["aws_internet_gateway.default"]
}

resource "aws_eip" "eip_nat_{{.az}}" {
  vpc = true
}

resource "aws_route" "route_private_{{.az}}" {
    route_table_id         = "${aws_route_table.route_table_private_{{.az}}.id}"
    destination_cidr_block = "0.0.0.0/0"
    nat_gateway_id         = "${aws_nat_gateway.nat_gateway_{{.az}}.id}"
}

resource "aws_route_table" "route_table_private_{{.az}}" {
    vpc_id = "${aws_vpc.default.id}"
}

resource "aws_route_table" "route_table_protected_{{.az}}" {
    vpc_id = "${aws_vpc.default.id}"
}
`
)

func Output(sn subnet.Subnet, extras *[]net.IPNet) (string, error) {
	var buf bytes.Buffer
	tmplPreamble, err := template.New("preamble").Parse(PREAMBLE)
	if err != nil {
		return "", err
	}

	tmplSubnet, err := template.New("subnet").Parse(SUBNET)
	if err != nil {
		return "", err
	}

	tmplRouting, err := template.New("routing").Parse(ROUTING)
	if err != nil {
		return "", err
	}

	tmplPreamble.Execute(&buf, map[string]string{"cidrblock": sn.VPC.String()})
	for k, v := range sn.AvailabilityZones {
		for _, t := range []block{{v.Public, "public"}, {v.Private, "private"}, {v.Protected, "protected"}} {
			tmplSubnet.Execute(&buf, map[string]string{
				"az":             strconv.Itoa(k),
				"cidrblockInner": t.Addr.String(),
				"function":       t.Type,
			})
		}

		tmplRouting.Execute(&buf, map[string]string{"az": strconv.Itoa(k)})
	}

	return buf.String(), nil
}
