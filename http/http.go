package http

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/tschuy/cidrblocks/output/table"
	"github.com/tschuy/cidrblocks/output/terraform"
	"github.com/tschuy/cidrblocks/subnet"
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

		sn := subnet.New(ipnet, 4, 4)

		var cidrOut string

		functions := map[string]func(subnet.Subnet) (string, error){
			"table":     table.Output,
			"terraform": terraform.Output,
		}

		if function, ok := functions[format]; ok {
			cidrOut, err = function(sn)
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
