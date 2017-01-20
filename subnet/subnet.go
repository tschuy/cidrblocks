package subnet

import (
	"math"
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
	Extra     net.IPNet
	AZBlock   net.IPNet
}

func AZName(k int) string {
	return string(k + 97) // we will never have more than 26 AZs
	// currently, the maximum number of AZs Amazon has in any region appears
	// to be around five-ish (us-east-1 goes from a-e)
	// if you're looking at this comment to try to add support for 27+, check
	// git history to get the answer for free.
}

func timesSplit(k int) int {
	num := math.Ceil(math.Log2(float64(k)))
	if num == 0 {
		return 1
	}
	return int(math.Pow(2, num))
}

func New(ipnet *net.IPNet, azs int) (*Subnet, *[]net.IPNet, error) {
	// split into the number of necessary blocks for AZs
	splits := timesSplit(azs)

	azblocks, err := ipnets.SubnetInto(ipnet, timesSplit(azs))
	if err != nil {
		return nil, nil, err
	}

	var subnet Subnet
	subnet.VPC = ipnet

	for az := 0; az < azs; az++ {
		halves, err := ipnets.SubnetInto(azblocks[az], 2) // private is half the total network
		if err != nil {
			return nil, nil, err
		}

		quarters, err := ipnets.SubnetInto(halves[1], 2) // public is a quarter of the total network
		if err != nil {
			return nil, nil, err
		}

		eighths, err := ipnets.SubnetInto(quarters[1], 2) // protected is an eighth of the total network
		if err != nil {
			return nil, nil, err
		}

		subnet.AvailabilityZones = append(subnet.AvailabilityZones, AvailabilityZone{
			AZBlock:   *azblocks[az],
			Private:   *halves[0],   // first half of AZBlock
			Public:    *quarters[0], // third quarter of AZBlock
			Protected: *eighths[0],  // seventh eighth of AZBlock
			Extra:     *eighths[1],  // last eighth of AZBlock is reserved for future use
		})
	}

	var extras []net.IPNet
	for t := azs; t < splits; t++ {
		extras = append(extras, *azblocks[t])
	}

	return &subnet, &extras, nil

}
