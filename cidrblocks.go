package main

// pretty print -- aka outputTable
// Terraform
// cloud formation
  // 1. create the VPC
  // 2. create internet gateway
  // 3. create NAT gateway [in one of the public subnets]
  // 4. create three route tables -- public inc. 0.0.0.0 towards internet gateway, private inc. 0.0.0.0 pointed towards NAT gateway, protected
  // to test -- fire off cloud formation, create the VPC, spin up instances in pub/priv and check they can connect to outside
      // to be super cool, make machines with cloud formation and then make a cloud config to curl and show it's working
// AWS commands -- aws ec2 create-subnet, etc

// SHOULD be useable as a not-web-service



import (
  "bytes"
  "fmt"
	"io"
  "log"
  "math"
  "net"
	"net/http"
  "strconv"
  "text/template"

  "github.com/mhuisi/ipv4utils"
)

func handle(w http.ResponseWriter, r *http.Request) {
  format := r.URL.Query().Get("format")
  cidr := r.URL.Query().Get("cidr")
  if format != "" && cidr != "" {
    // parse cidr config
    _, ipnet, err := net.ParseCIDR(cidr)

    if (err != nil) {
      http.Error(w, fmt.Sprintf(`{"error": %s}`, err), http.StatusBadRequest)
      return
    }

    block := divideSubnets(ipnet, 4)
    vpccidr, allocated := allocateSubnets(block, 4)

    var cidrOut string
    if format == "table" {
      cidrOut, err = outputTable(vpccidr, allocated)
    } else if format == "terraform" {
     cidrOut, err = outputTerraform(vpccidr, allocated)
    } else {
      http.Error(w, fmt.Sprintf(`{"error": format %s not recognized}`, format), http.StatusBadRequest)
      return
    }

    if (err != nil) {
      http.Error(w, fmt.Sprintf(`{"error": %s}`, err), http.StatusBadRequest)
      return
    }

    io.WriteString(w, cidrOut)
    return

  } else {
    http.Error(w, `{"error": "missing parameter (need format and cidr)"}`, http.StatusBadRequest)
    return
  }
}

func main() {
	http.HandleFunc("/", handle)
	err := http.ListenAndServe(":8087", nil)
  if err != nil {
    panic(err)
  }
}

func outputTerraform(vpccidr *net.IPNet, alloc []map[string]*net.IPNet) (string, error) {
  var buf bytes.Buffer
  tmplPreamble, err := template.New("preamble").Parse(`variable "cidr_block" {
    type = "string"
    default = "{{.cidrblock}}"
}

# Specify the provider and access details
provider "aws" {

}

data "aws_region" "default" {
  current = true
}

# Create a VPC to launch our instances into
resource "aws_vpc" "default" {
    cidr_block = "${var.cidr_block}"
    enable_dns_hostnames = true
}

# Grant the VPC internet access on its main route table
resource "aws_route" "internet_access" {
    route_table_id         = "${aws_vpc.default.main_route_table_id}"
    destination_cidr_block = "0.0.0.0/0"
    gateway_id             = "${aws_internet_gateway.default.id}"
}

resource "aws_internet_gateway" "default" {
  vpc_id = "${aws_vpc.default.id}"

    tags {
            Name = "vpc-igw"
          }
}`)

  if err != nil {
    return "", err
  }

  tmplAZ, err := template.New("az").Parse(`

resource "aws_subnet" "AZ-{{.az}}-{{.function}}" {
vpc_id                  = "${aws_vpc.default.id}"
cidr_block              = "{{.cidrblockInner}}"
availability_zone       = "${data.aws_region.default.name}{{.az}}"
map_public_ip_on_launch = false
}`)

  if err != nil {
    return "", err
  }

  tmplPreamble.Execute(&buf, map[string]string{"cidrblock": vpccidr.String()})
  for k, v := range alloc {
    for _, t := range []string{"public", "private", "protected"} {
      tmplAZ.Execute(&buf, map[string]string{
        "az": string(k + 65),
        "cidrblockInner": v[t].String(),
        "function": t,
      })
    }
  }

  return buf.String(), nil
}

func outputTable(vpccidr *net.IPNet, alloc []map[string]*net.IPNet) (string, error) {

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
    if err != nil { panic(err) }
    infomap := make(map[string]string)
    infomap["az"] = string(k + 65)

    infomap["azblock"] = v["azblock"].String()

    infomap["private"] = v["private"].String()
    a, b := v["private"].Mask.Size()
    infomap["privcount"] = strconv.FormatFloat(math.Pow(2, float64(b - a)), 'f', 0, 64)

    infomap["public"] = v["public"].String()
    a, b = v["public"].Mask.Size()
    infomap["pubcount"] = strconv.FormatFloat(math.Pow(2, float64(b - a)), 'f', 0, 64)

    infomap["protected"] = v["protected"].String()
    a, b = v["protected"].Mask.Size()
    infomap["protcount"] = strconv.FormatFloat(math.Pow(2, float64(b - a)), 'f', 0, 64)


    var tbl bytes.Buffer
    err = tmpl.Execute(&tbl, infomap)
    if err != nil {
      return "", err
    }

    buf.WriteString(tbl.String())
  }
  return buf.String(), nil
}

func allocateSubnets(block map[string][]*net.IPNet, azs int) (*net.IPNet, []map[string]*net.IPNet) {
  var allocated []map[string]*net.IPNet // map to slices of IPNets
  vpccidr := block["l1"][0]
  for az := 0; az < azs; az++ {
    allocated = append(allocated, map[string]*net.IPNet{
      "azblock": block["l2"][0 + az * 1],
      "private": block["l3"][0 + az * 2],
      "public":  block["l4"][2 + az * 4],
      "protected": block["l5"][6 + az * 8],
    })
  }

  return vpccidr, allocated
}

func divideSubnets(ipnet *net.IPNet, depth int) (map[string][]*net.IPNet) {
  block := make(map[string][]*net.IPNet) // map to slices of IPNets
  _, _ = ipnet.Mask.Size() // TODO subnet

  block["l1"] = []*net.IPNet{ipnet}

  for i := 0; i < depth; i++ {
  	lvl := fmt.Sprintf("l%d", i+2)
    schan, err := ipv4utils.Subnet(*ipnet, uint(i + 2))
    if err != nil {
      log.Fatal(err)
    }
    for subnet := range schan {
      // make a copy of subnet, because memory reuse :(
      var dupSubnet net.IPNet

      dupIp := make(net.IP, len(subnet.IP))
      copy(dupIp, subnet.IP)
      dupSubnet.IP = dupIp

      dupMask := make(net.IPMask, len(subnet.Mask))
      copy(dupMask, subnet.Mask)
      dupSubnet.Mask = dupMask
      block[lvl] = append(block[lvl], &dupSubnet)
    }
	}

  return block
}
