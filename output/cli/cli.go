package cli

import (
	"bytes"
	"net"
	"strconv"
	"text/template"

	"github.com/tschuy/cidrblocks/subnet"
)

const TMPLSTRING = `export VPCCIDR="10.0.0.0/20"
export AWS_PROFILE="coreos-dev"
export AWS_DEFAULT_REGION="us-east-1"
# get array of available AZs
AZS=$(aws ec2 describe-availability-zones | jq -r '.AvailabilityZones[].ZoneName')
IFS=', ' read -r -a AZS <<< $AZS

VPCID=$(aws ec2 create-vpc --cidr-block $VPCCIDR | jq -r .Vpc.VpcId)
if [[ ${PIPESTATUS[0]} -ne 0 ]] ; then
  echo "Failed to create vpc!"
  exit 1
fi

echo "Created VPC $VPCID"

IGWID=$(aws ec2 create-internet-gateway | jq -r .InternetGateway.InternetGatewayId)
if [[ ${PIPESTATUS[0]} -ne 0 ]] ; then
  echo "Failed to create internet gateway!"
  exit 1
fi

echo "Created internet gateway $IGWID"

IRTB=$(aws ec2 create-route-table --vpc-id $VPCID | jq -r .RouteTable.RouteTableId)
# ERR
echo "Created internet route table $IRTB"

aws ec2 attach-internet-gateway --internet-gateway-id $IGWID --vpc-id $VPCID
# ERR
echo "Attached internet gateway to VPC"

aws ec2 create-route --route-table-id $IRTB --gateway-id $IGWID --destination-cidr-block 0.0.0.0/0
# ERR
echo "Created route on route table to internet gateway"

PUB_CIDRS=({{.pubs}})
PRIV_CIDRS=({{.privs}})
PROT_CIDRS=({{.prots}})
NUM_AZS={{.num}}

for ((CURR_AZ=0; CURR_AZ < $NUM_AZS; CURR_AZ++)); do
  # loop this bit for each AZ
  PUB_SUBNET=$(aws ec2 create-subnet --vpc-id $VPCID --cidr-block ${PUB_CIDRS[$CURR_AZ]} --availability-zone ${AZS[$CURR_AZ]} | jq -r .Subnet.SubnetId)
  # ERR
  aws ec2 modify-subnet-attribute --subnet-id $PUB_SUBNET --map-public-ip-on-launch
  # ERR - if output is not nothing it's an error

  PRIV_SUBNET=$(aws ec2 create-subnet --vpc-id $VPCID --cidr-block ${PRIV_CIDRS[$CURR_AZ]} --availability-zone ${AZS[$CURR_AZ]} | jq -r .Subnet.SubnetId)
  # ERR
  PROT_SUBNET=$(aws ec2 create-subnet --vpc-id $VPCID --cidr-block ${PROT_CIDRS[$CURR_AZ]} --availability-zone ${AZS[$CURR_AZ]} | jq -r .Subnet.SubnetId)
  # ERR

  ALLOC_ID=$(aws ec2 allocate-address --domain vpc | jq -r .AllocationId)
  # ERR
  NATGATEWAY=$(aws ec2 create-nat-gateway --subnet-id $PUB_SUBNET --allocation-id $ALLOC_ID | jq -r .NatGateway.NatGatewayId)
  # ERR

  PRIV_RT=$(aws ec2 create-route-table --vpc-id $VPCID | jq -r .RouteTable.RouteTableId)
  # ERR
  aws ec2 associate-route-table --subnet-id $PRIV_SUBNET --route-table-id $PRIV_RT
  # ERR

  aws ec2 create-route --route-table-id $PRIV_RT --gateway-id $NATGATEWAY --destination-cidr-block 0.0.0.0/0
  # ERR

  PROT_RT=$(aws ec2 create-route-table --vpc-id $VPCID | jq -r .RouteTable.RouteTableId)
  # ERR
  aws ec2 associate-route-table --subnet-id $PROT_SUBNET --route-table-id $PROT_RT
  # ERR

  aws ec2 associate-route-table --subnet-id $PUB_SUBNET --route-table-id $IRTB
  # ERR
done
`

func Output(sn subnet.Subnet, extras *[]net.IPNet) (string, error) {
	var out, pubs, privs, prots bytes.Buffer

	tmplBash, err := template.New("bash").Parse(TMPLSTRING)
	if err != nil {
		return "", err
	}

	for _, v := range sn.AvailabilityZones {
		pubs.WriteString(v.Public.String())
		pubs.WriteString(" ")
		privs.WriteString(v.Private.String())
		privs.WriteString(" ")
		prots.WriteString(v.Protected.String())
		prots.WriteString(" ")
	}

	tmplBash.Execute(&out, map[string]string{
		"pubs":  pubs.String(),
		"privs": privs.String(),
		"prots": prots.String(),
		"num":   strconv.Itoa(len(sn.AvailabilityZones)),
	})

	return out.String(), nil
}
