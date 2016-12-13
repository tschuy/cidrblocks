package main

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tschuy/cidrblocks/http"
)

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

var azs int
var port int
var cidr string
var format string
var rootCmd *cobra.Command

func main() {

	rootCmd = &cobra.Command{
		Long: `Split a CIDR block into availability zones and public/private/protected.

--cidr flag required.`,
		Run: cli,
	}
	rootCmd.Flags().StringVarP(&cidr, "cidr", "c", "required", "[required] root cidr block (ex: 10.0.0.0/8)")
	rootCmd.Flags().IntVarP(&azs, "azs", "a", 4, "number of availability zones (power of two)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "table", "format of output (table or terraform)")

	serveCmd := &cobra.Command{
		Use:  "serve",
		Long: `Start an HTTP server serving over port specified with --port`,
		Run: func(cmd *cobra.Command, args []string) {
			http.Serve(":" + strconv.Itoa(port))
		},
	}

	serveCmd.Flags().IntVarP(&port, "port", "p", 8087, "port to serve on")

	rootCmd.AddCommand(serveCmd)
	rootCmd.Execute()

}
