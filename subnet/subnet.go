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
	AZBlock   net.IPNet
}

func AZName(k int) string {
	// for the ever-present fear that you might suddenly use more than 26
	// availability zones!
	var char int
	// for 26, = 1, for 27, = 2
	// log(a)/log(b) = log base b of a
	// start at 1 instead of 0
	strLen := int(math.Floor(math.Log(float64(k))/math.Log(26) + 1))
	if strLen < 0 {
		// unfortunately, log(0) = -inf
		// we can't just add one to the log, as this would make k=26 have strLen
		// of 2, and not 1
		strLen = 1
	}
	name := make([]byte, strLen)
	k = k + 1 // so we start at A and not at space
	for i := strLen; i > 0; i-- {
		k = k - 1
		char = k % 26
		k = k / 26
		name[i-1] = byte(char + 65)
	}
	return string(name)
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
