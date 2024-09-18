package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	req := SearchRequest{}

	flag.IntVar(&req.Limit, "limit", 1, "")
	flag.IntVar(&req.Offset, "offset", 0, "")
	flag.StringVar(&req.Query, "query", "", "")
	flag.StringVar(&req.OrderField, "order-field", "", "")
	flag.IntVar(&req.OrderBy, "order-by", 0, "")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(flag.CommandLine.Output(), "file name required")
		flag.Usage()
		os.Exit(1)
	}

	dataset, err := LoadDatasetFromFile(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

	users := dataset.Search(req)
	if err != nil {
		log.Fatal(err)
	}

	buf, err := json.MarshalIndent(users, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout.Write(buf)
	os.Stdout.Write([]byte("\n"))
}
