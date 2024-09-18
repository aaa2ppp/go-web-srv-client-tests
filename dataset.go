package main

import (
	"encoding/xml"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type DatasetRow struct {
	ID        uint64 `xml:"id" json:"id"`
	FirstName string `xml:"first_name" json:"-"`
	LastName  string `xml:"last_name" json:"-"`
	Name      string `json:"name,omitempty"`
	Age       uint   `xml:"age" json:"age,omitempty"`
	Gender    string `xml:"gender" json:"gender,omitempty"`
	About     string `xml:"about" json:"about,omitempty"`
}

type Dataset []*DatasetRow

func LoadDataset(r io.Reader) (Dataset, error) {
	var root struct {
		Rows Dataset `xml:"row"`
	}

	if err := xml.NewDecoder(r).Decode(&root); err != nil {
		return nil, err
	}

	for _, row := range root.Rows {
		row.Name = row.FirstName + " " + row.LastName
	}

	return root.Rows, nil
}

func LoadDatasetFromFile(path string) (Dataset, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return LoadDataset(file)
}

func (ds Dataset) Search(req SearchRequest) []*DatasetRow {
	const op = "Dataset.Search"

	log.Printf("%s: req: %+v", op, req)

	res := ds.filter(req.Query)
	log.Printf("%s: filtered %d records from %d", op, len(res), len(ds))

	if req.Offset >= len(res) {
		log.Printf("%s: offset=%d out of range, return empty list", op, req.Offset)
		return []*DatasetRow{} // empty list
	}

	if req.OrderBy != 0 {
		res.sort(req.OrderField)
	}

	if req.OrderBy < 0 {
		res.reverse()
	}

	end := min(req.Offset+req.Limit, len(res))
	res = res[req.Offset:end]

	log.Printf("%s: return %d records", op, len(res))
	return res
}

func (ds Dataset) filter(query string) Dataset {
	var res Dataset

	if query == "" {
		res = make(Dataset, len(ds))
		copy(res, ds)
		return res
	}

	for _, u := range ds {
		if strings.Contains(u.Name, query) || strings.Contains(u.About, query) {
			res = append(res, u)
		}
	}

	return res
}

func (ds Dataset) sort(field string) {
	const op = "Dataset.sort"

	switch strings.ToLower(field) {
	case "name", "":
		log.Printf("%s: by name", op)
		sort.Slice(ds, func(i, j int) bool {
			return ds[i].Name < ds[j].Name
		})
	case "id":
		log.Printf("%s: by id", op)
		sort.Slice(ds, func(i, j int) bool {
			return ds[i].ID < ds[j].ID
		})
	case "age":
		log.Printf("%s: by age", op)
		sort.Slice(ds, func(i, j int) bool {
			return ds[i].Age < ds[j].Age
		})
	}
}

func (ds Dataset) reverse() {
	log.Println("Dataset.reverse")
	for i, j := 0, len(ds)-1; i < j; i, j = i+1, j-1 {
		ds[i], ds[j] = ds[j], ds[i]
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
