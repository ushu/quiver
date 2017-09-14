/*
The quiver_to_json tool loads a provided Quiver library into a single JSON.

Usage:

	# To load all the lib contents into a single JSON
	$ quiver_to_json /path/to/Quiver.qvlibrary > quiver.json

	# To include the content of all resources as data URIs
	$ quiver_to_json -res /path/to/Quiver.qvlibrary > quiver.json
*/
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"flag"

	"github.com/ushu/quiver"
)

// Tells the tool to also load the resources.
var flagRes bool

func init() {
	flag.BoolVar(&flagRes, "res", false, "load resources in JSON")
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Usage: quiver_to_json [-res] QUIVER_LIBRARY")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read full library into memory
	library, err := quiver.ReadLibrary(flag.Args()[0], flagRes)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Outputs the library as JSON
	err = json.NewEncoder(os.Stdout).Encode(library)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
