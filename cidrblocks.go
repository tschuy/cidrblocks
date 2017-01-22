package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tschuy/cidrblocks/http"
	awscli "github.com/tschuy/cidrblocks/output/cli"
	"github.com/tschuy/cidrblocks/output/cloudformation"
	"github.com/tschuy/cidrblocks/output/table"
	"github.com/tschuy/cidrblocks/output/terraform"
	"github.com/tschuy/cidrblocks/subnet"
)

var azs int
var port int
var cidr string
var format string
var rootCmd *cobra.Command

func main() {

	rootCmd = &cobra.Command{
		Long: `Split a CIDR block into availability zones and public/private/protected.`,
		Run:  cli,
	}

	red := color.New(color.Bold).SprintFunc()

	rootCmd.Flags().StringVarP(&cidr, "cidr", "c", "", red("[required] root cidr block (ex: 10.0.0.0/8)"))
	rootCmd.Flags().IntVarP(&azs, "azs", "a", 4, "number of availability zones (power of two)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "table", "format of output (table | terraform | cloudformation)")

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

func cli(cmd *cobra.Command, args []string) {
	if cidr == "" {
		// empty string is the default
		help := rootCmd.HelpFunc()
		help(rootCmd, []string{})

		os.Exit(1)
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sn, extras, err := subnet.New(ipnet, azs)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	functions := map[string]func(subnet.Subnet, *[]net.IPNet) (string, error){
		"table":          table.Output,
		"terraform":      terraform.Output,
		"cloudformation": cloudformation.Output,
		"cli":            awscli.Output,
	}

	var cidrOut string

	if function, ok := functions[format]; ok {
		cidrOut, err = function(*sn, extras)
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
