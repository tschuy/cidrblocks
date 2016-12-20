package subnet

import (
	"net"

	"github.com/dghubble/ipnets"
)

type Subnet struct {
	AvailabilityZones []AvailabilityZone
	VPC               *net.IPNet
}

type AvailabilityZone struct {
	Public    net.IPNet
	Private   net.IPNet
	Protected net.IPNet
	AZBlock   net.IPNet
}

func AZName(k int) string {
	return string(char + 97) // we will never have more than 26 AZs
	// currently, the maximum number of AZs Amazon has in any region appears
	// to be around five-ish (us-east-1 goes from a-e)
	// if you're looking at this comment to try to add support for 27+, check
	// git history to get the answer for free.
}

func New(ipnet *net.IPNet, azs int) (*Subnet, error) {
	// split into the number of necessary blocks for AZs
	azblocks, err := ipnets.SubnetInto(*ipnet, azs)
	if err != nil {
		return nil, err
	}

	var subnet Subnet
	subnet.VPC = ipnet

	for az := 0; az < azs; az++ {
		halves, err := ipnets.SubnetInto(azblocks[az], 2) // private is half the total network
		if err != nil {
			return nil, err
		}

		quarters, err := ipnets.SubnetInto(halves[1], 2) // public is quarter the total network
		if err != nil {
			return nil, err
		}

		eighths, err := ipnets.SubnetInto(quarters[1], 2) // protected is quarter the total network
		if err != nil {
			return nil, err
		}

		subnet.AvailabilityZones = append(subnet.AvailabilityZones, AvailabilityZone{
			AZBlock:   azblocks[az],
			Private:   halves[0],   // first half of AZBlock
			Public:    quarters[0], // third quarter of AZBlock
			Protected: eighths[0],  // seventh eighth of AZBlock
			// last eighth is unused
		})
	}

	return &subnet, nil

}
