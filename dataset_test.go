package main

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestLoadDataset(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    Dataset
		wantErr bool
	}{
		{
			"1",
			args{strings.NewReader(`<?xml version="1.0" encoding="UTF-8" ?>
<root>
  <row>
    <id>0</id>
    <guid>1a6fa827-62f1-45f6-b579-aaead2b47169</guid>
    <age>22</age>
    <first_name>Boyd</first_name>
    <last_name>Wolf</last_name>
    <gender>male</gender>
    <about>some text...</about>
  </row>
  <row>
    <id>1</id>
    <guid>46c06b5e-dd08-4e26-bf85-b15d280e5e07</guid>
    <age>21</age>
    <first_name>Hilda</first_name>
    <last_name>Mayer</last_name>
    <gender>female</gender>
    <about>some text 2...</about>
  </row>
</root>
`)},
			Dataset{
				{
					ID:        0,
					Age:       22,
					FirstName: "Boyd",
					LastName:  "Wolf",
					Name:      "Boyd Wolf",
					Gender:    "male",
					About:     "some text...",
				},
				{
					ID:        1,
					Age:       21,
					FirstName: "Hilda",
					LastName:  "Mayer",
					Name:      "Hilda Mayer",
					Gender:    "female",
					About:     "some text 2...",
				},
			},
			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := LoadDataset(tt.args.r)
		if (err != nil) != tt.wantErr {
			t.Errorf("ReadFromXML() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("ReadFromXML() = %v, want %v", got, tt.want)
		}
	}
}
