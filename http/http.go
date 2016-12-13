package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/tschuy/cidrblocks/output/table"
	"github.com/tschuy/cidrblocks/output/terraform"
	"github.com/tschuy/cidrblocks/subnet"
)

type param struct {
	format string
	ipnet  *net.IPNet
	azs    int
	depth  int
}

type JsonError struct {
	Err string `json:"error"`
}

func Serve(address string) {
	http.HandleFunc("/", handle)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}

func parseParams(r *http.Request) (*param, error) {
	var params param
	var n int

	format := r.URL.Query().Get("format")
	if format == "" {
		return nil, errors.New("parameter format cannot be empty")
	}
	params.format = format

	cidrParam := r.URL.Query().Get("cidr")
	_, ipnet, err := net.ParseCIDR(cidrParam)
	if err != nil {
		return nil, err
	}
	params.ipnet = ipnet

	azs := r.URL.Query().Get("azs")
	if azs != "" {
		n, err = strconv.Atoi(azs)
		if err != nil {
			return nil, err
		}
	} else {
		n = 4 // default availability zones
	}
	params.azs = n

	return &params, nil
}

func handle(w http.ResponseWriter, r *http.Request) {
	params, err := parseParams(r)

	if err != nil {
		errText, _ := json.Marshal(JsonError{err.Error()})
		http.Error(w, string(errText), http.StatusBadRequest)
		return
	}
	sn, err := subnet.New(params.ipnet, params.azs)

	if err != nil {
		errText, _ := json.Marshal(JsonError{err.Error()})
		http.Error(w, string(errText), http.StatusBadRequest)
		return
	}

	var cidrOut string

	functions := map[string]func(subnet.Subnet) (string, error){
		"table":     table.Output,
		"terraform": terraform.Output,
	}

	if function, ok := functions[params.format]; ok {
		cidrOut, err = function(*sn)
	} else {
		errText, _ := json.Marshal(JsonError{fmt.Sprintf("format %s not recognized", params.format)})
		http.Error(w, string(errText), http.StatusBadRequest)
		return
	}

	if err != nil {
		errText, _ := json.Marshal(JsonError{err.Error()})
		http.Error(w, string(errText), http.StatusBadRequest)
		return
	}

	io.WriteString(w, cidrOut)
	return

}
