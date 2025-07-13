package coredns_mikrotik_dhcp

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func writeJSONResponse(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

func TestMikroTikAPILeaseGetter(t *testing.T) {
	expectedLeases := []Lease{
		{
			Address:  net.ParseIP("192.168.10.1"),
			Hostname: "host1",
		},
		{
			Address:  net.ParseIP("192.168.10.2"),
			Hostname: "host2",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /rest/ip/dhcp-server/lease", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != "foobar" {
			resp := errorResponse{Error: http.StatusUnauthorized, Message: "Unauthorized"}
			writeJSONResponse(w, http.StatusUnauthorized, resp)
			return
		}

		if r.URL.Query().Get("status") != "bound" {
			// ensure query parameter to filter status is set
			resp := errorResponse{Error: http.StatusBadRequest, Message: "Bad Request"}
			writeJSONResponse(w, http.StatusUnauthorized, resp)
			return
		}

		writeJSONResponse(w, http.StatusOK, expectedLeases)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := errorResponse{Error: http.StatusBadRequest, Message: "Bad Request"}
		writeJSONResponse(w, http.StatusUnauthorized, resp)
		return
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Successfully get bound leases", func(t *testing.T) {
		lg := NewMikroTikAPILeaseGetter(ts.URL, "admin", "foobar")
		actualLeases, err := lg.GetBoundLeases(ctx)
		if err != nil {
			t.Fatalf("expected err to be nil, got %v", err)
		}

		if !reflect.DeepEqual(actualLeases, expectedLeases) {
			t.Fatalf("expected leases to be %v, got %v", expectedLeases, actualLeases)
		}
	})

	t.Run("Authorization error", func(t *testing.T) {
		lg := NewMikroTikAPILeaseGetter(ts.URL, "admin", "invalid")
		actualLeases, err := lg.GetBoundLeases(ctx)
		if err == nil {
			t.Fatalf("expected error to occur, got nil")
		}

		if actualLeases != nil {
			t.Fatalf("expected leases to be nil, got %v", actualLeases)
		}
	})
}
