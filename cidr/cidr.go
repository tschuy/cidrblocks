package cidr

import (
	"fmt"
	"log"
	"net"

	"github.com/mhuisi/ipv4utils"
)

func AllocateSubnets(block map[string][]*net.IPNet, azs int) (*net.IPNet, []map[string]*net.IPNet) {
	var allocated []map[string]*net.IPNet // map to slices of IPNets
	vpccidr := block["l1"][0]
	for az := 0; az < azs; az++ {
		allocated = append(allocated, map[string]*net.IPNet{
			"azblock":   block["l2"][0+az*1],
			"private":   block["l3"][0+az*2],
			"public":    block["l4"][2+az*4],
			"protected": block["l5"][6+az*8],
		})
	}

	return vpccidr, allocated
}

func DivideSubnets(ipnet *net.IPNet, depth int) map[string][]*net.IPNet {
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
