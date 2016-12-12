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

import "github.com/tschuy/cidrblocks/http"

func main() {
	http.Serve()
}
