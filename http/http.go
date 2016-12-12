package http

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/tschuy/cidrblocks/cidr"
	"github.com/tschuy/cidrblocks/output/table"
	"github.com/tschuy/cidrblocks/output/terraform"
)

func Serve() {
	http.HandleFunc("/", handle)
	err := http.ListenAndServe(":8087", nil)
	if err != nil {
		panic(err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	cidrParam := r.URL.Query().Get("cidr")
	if format != "" && cidrParam != "" {
		// parse cidr config
		_, ipnet, err := net.ParseCIDR(cidrParam)

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": %s}`, err), http.StatusBadRequest)
			return
		}

		block := cidr.DivideSubnets(ipnet, 4)
		vpccidr, allocated := cidr.AllocateSubnets(block, 4)

		var cidrOut string

		functions := map[string]func(vpccidr *net.IPNet, alloc []map[string]*net.IPNet) (string, error){
			"table":     table.Output,
			"terraform": terraform.Output,
		}

		if function, ok := functions[format]; ok {
			cidrOut, err = function(vpccidr, allocated)
		} else {
			http.Error(w, fmt.Sprintf(`{"error": format %s not recognized}`, format), http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": %s}`, err), http.StatusBadRequest)
			return
		}

		io.WriteString(w, cidrOut)
		return

	} else {
		http.Error(w, `{"error": "missing parameter (need format and cidr)"}`, http.StatusBadRequest)
		return
	}
}
