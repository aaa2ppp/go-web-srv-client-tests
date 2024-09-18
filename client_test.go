package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// код писать тут

const datasetPath = "./dataset.xml"

var dataset Dataset

func init() {
	const op = "init"

	var err error
	dataset, err = LoadDatasetFromFile(datasetPath)
	if err != nil {
		panic(err)
	}

	log.Printf("%s: load %d records", op, len(dataset))
}

type SearchServer struct {
	dataset Dataset
	req     SearchRequest
	resp    Dataset
}

func writeResponse(w http.ResponseWriter, status int, resp interface{}) {
	const op = "writeResponse"

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func writeBadRequestResponse(w http.ResponseWriter, errorMsg string) {
	writeResponse(w, http.StatusBadRequest, &SearchErrorResponse{Error: errorMsg})
}

const (
	errorBadLimit      = "ErrorBadLimit"
	errorBadOffset     = "ErrorBadOffset"
	errorBadOrderField = "ErrorBadOrderField"
	errorBadOrderBy    = "ErrorBadOrderBy"
)

func (h *SearchServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "SearchServer.ServeHTTP"

	req := SearchRequest{
		Limit:      1,
		OrderField: "name",
	}

	log.Printf("%s: url: %v", op, r.URL)

	q := r.URL.Query()

	if s := q.Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 {
			log.Printf("%s: %s: %s", op, errorBadLimit, s)
			writeBadRequestResponse(w, errorBadLimit)
			return
		}
		req.Limit = v
	}

	if s := q.Get("offset"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			log.Printf("%s: %s: %s", op, errorBadOffset, s)
			writeBadRequestResponse(w, errorBadOffset)
			return
		}
		req.Offset = v
	}

	if s := q.Get("query"); s != "" {
		req.Query = s
	}

	if s := q.Get("order_field"); s != "" {
		switch s {
		case "Id", "Age", "Name": // ok
		default:
			log.Printf("%s: %s: %s", op, errorBadOrderField, s)
			writeBadRequestResponse(w, errorBadOrderField)
			return
		}
		req.OrderField = s
	}

	if s := q.Get("order_by"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || !(v == -1 || v == 0 || v == 1) {
			log.Printf("%s: %s: %s", op, errorBadOrderBy, s)
			writeBadRequestResponse(w, errorBadOrderBy)
			return
		}
		req.OrderBy = v
	}

	h.req = req
	h.resp = h.dataset.Search(req)
	writeResponse(w, http.StatusOK, h.resp)
}

type DummyServer struct {
	status     int
	timeout    time.Duration
	serverFail bool
}

func (h *DummyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.serverFail {
		panic("*** crash test ***")
	}
	if h.timeout != 0 {
		time.Sleep(h.timeout)
	}
	w.WriteHeader(h.status)
}

func getWantResponse(ss *SearchServer) *SearchResponse {
	users := make([]User, len(ss.resp))
	for i, u := range ss.resp {
		users[i] = User{
			Id:     int(u.ID),
			Name:   u.Name,
			Age:    int(u.Age),
			Gender: u.Gender,
			About:  u.About,
		}
	}

	nextPage := len(users) == ss.req.Limit
	if nextPage {
		users = users[:len(users)-1]
	}

	return &SearchResponse{Users: users, NextPage: nextPage}
}

func TestSearchClient_FindUsers(t *testing.T) {
	tests := []struct {
		name    string
		server  http.Handler
		req     SearchRequest
		want    *SearchResponse
		wantErr bool
	}{
		{
			"bad limit",
			&DummyServer{status: http.StatusOK},
			SearchRequest{Limit: -1},
			nil,
			true,
		},
		{
			"bad offset",
			&DummyServer{status: http.StatusOK},
			SearchRequest{Offset: -1},
			nil,
			true,
		},
		{
			"unauthorized",
			&DummyServer{status: http.StatusUnauthorized},
			SearchRequest{},
			nil,
			true,
		},
		{
			"bad request",
			&DummyServer{status: http.StatusBadRequest},
			SearchRequest{},
			nil,
			true,
		},
		{
			"no body",
			&DummyServer{status: http.StatusOK},
			SearchRequest{},
			nil,
			true,
		},
		{
			"internal error",
			&DummyServer{status: http.StatusInternalServerError},
			SearchRequest{},
			nil,
			true,
		},
		{
			"client timeout + 100ms",
			&DummyServer{status: http.StatusOK, timeout: client.Timeout + 100*time.Millisecond},
			SearchRequest{},
			nil,
			true,
		},
		{
			"server fail",
			&DummyServer{status: http.StatusOK, serverFail: true},
			SearchRequest{},
			nil,
			true,
		},
		{
			"bad request: " + errorBadOrderField,
			&SearchServer{dataset: dataset},
			SearchRequest{OrderField: "jlcGd0NJ"},
			nil,
			true,
		},
		{
			"bad request: " + errorBadOrderBy,
			&SearchServer{dataset: dataset},
			SearchRequest{OrderBy: 100500},
			nil,
			true,
		},
		{
			"limit=1",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 1},
			nil, // getWantLimit
			false,
		},
		{
			"limit=100500",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 100500},
			nil, // getWantLimit
			false,
		},
		{
			"limit=100500,query=1ZeE6BqL",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 100500, Query: "1ZeE6BqL"},
			nil, // getWantLimit
			false,
		},
		{
			"limit=100500,order_by=1",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 100500, OrderBy: 1},
			nil, // getWantLimit
			false,
		},
		{
			"limit=100500,order_by=-1",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 100500, OrderBy: -1},
			nil, // getWantLimit
			false,
		},
		{
			"limit=100500,order_field=Name,order_by=1",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 100500, OrderField: "Name", OrderBy: 1},
			nil, // getWantLimit
			false,
		},
		{
			"limit=100500,order_field=Id,order_by=1",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 100500, OrderField: "Id", OrderBy: 1},
			nil, // getWantLimit
			false,
		},
		{
			"limit=100500,order_field=Age,order_by=1",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 100500, OrderField: "Age", OrderBy: 1},
			nil, // getWantLimit
			false,
		},
		{
			"limit=100500,query=Velit",
			&SearchServer{dataset: dataset},
			SearchRequest{Limit: 100500, Query: "Velit"},
			nil, // getWantLimit
			false,
		},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.server)
			defer ts.Close()

			srv := &SearchClient{URL: ts.URL}

			got, err := srv.FindUsers(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchClient.FindUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				t.Logf("SearchClient.FindUsers() expected error: %v", err)
			}

			if ss, ok := tt.server.(*SearchServer); ok && err == nil {
				tt.want = getWantResponse(ss)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchClient.FindUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}
