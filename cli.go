package main

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/tschuy/cidrblocks/output/table"
	"github.com/tschuy/cidrblocks/output/terraform"
	"github.com/tschuy/cidrblocks/subnet"
)

func cli(cmd *cobra.Command, args []string) {
	if cidr == "required" {
		// this is the default cidr string
		fmt.Println("Error: --cidr argument required\n")

		help := rootCmd.HelpFunc()
		help(rootCmd, []string{})

		os.Exit(1)
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sn, err := subnet.New(ipnet, azs)

	if err != nil {
		fmt.Println(err)
	}

	functions := map[string]func(subnet.Subnet) (string, error){
		"table":     table.Output,
		"terraform": terraform.Output,
	}

	var cidrOut string

	if function, ok := functions[format]; ok {
		cidrOut, err = function(*sn)
	} else {
		fmt.Println(fmt.Sprintf("format %s not recognized", format))
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(cidrOut)
}
