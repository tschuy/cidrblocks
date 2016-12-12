package subnet

import (
	"fmt"
	"log"
	"net"

	"github.com/mhuisi/ipv4utils"
)

type Subnet struct {
	AvailabilityZones []AvailabilityZone
	VPC               *net.IPNet
}

type AvailabilityZone struct {
	Public    *net.IPNet
	Private   *net.IPNet
	Protected *net.IPNet
	AZBlock   *net.IPNet
}

func New(ipnet *net.IPNet, depth int, azs int) Subnet {
	block := divideSubnets(ipnet, depth)

	var subnet Subnet
	subnet.VPC = block["l1"][0]

	for az := 0; az < azs; az++ {
		subnet.AvailabilityZones = append(subnet.AvailabilityZones, AvailabilityZone{
			AZBlock:   block["l2"][0+az*1],
			Private:   block["l3"][0+az*2],
			Public:    block["l4"][2+az*4],
			Protected: block["l5"][6+az*8],
		})
	}

	return subnet
}

func divideSubnets(ipnet *net.IPNet, depth int) map[string][]*net.IPNet {
	block := make(map[string][]*net.IPNet) // map to slices of IPNets
	_, _ = ipnet.Mask.Size()               // TODO subnet

	block["l1"] = []*net.IPNet{ipnet}

	for i := 0; i < depth; i++ {
		lvl := fmt.Sprintf("l%d", i+2)
		schan, err := ipv4utils.Subnet(*ipnet, uint(i+2))
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
