package diskv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "github.com/micromdm/nanomdm/http/api"
	"github.com/micromdm/nanomdm/mdm"

	"github.com/micromdm/nanolib/log"
)

const deviceInfoReport = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CommandUUID</key><string>c1</string>
<key>QueryResponses</key><dict>
<key>iTunesStoreAccountIsActive</key><true/>
<key>iTunesStoreAccountHash</key><string>ctqKBqg9</string>
<key>OSVersion</key><string>27.2</string>
</dict>
<key>Status</key><string>Acknowledged</string>
<key>UDID</key><string>dev1</string>
</dict></plist>`

func TestCommandReportAPI(t *testing.T) {
	s := New(t.TempDir())
	ctx := context.Background()
	r := mdm.NewRequestWithContext(ctx, nil)
	r.EnrollID = &mdm.EnrollID{Type: mdm.Device, ID: "dev1"}

	c := &mdm.Command{CommandUUID: "c1", Raw: []byte("<plist/>")}
	c.Command.RequestType = "DeviceInformation"
	if _, err := s.EnqueueCommand(ctx, []string{"dev1"}, c); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	httpapi.HandleAPIv1("/v1", mux, log.NopLogger, s, nil)
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w
	}

	if w := get("/v1/report/dev1/c1"); w.Code != http.StatusNotFound || w.Header().Get("Content-type") != "application/json" {
		t.Fatalf("before the device reported: got %d, want 404", w.Code)
	}

	err := s.StoreCommandReport(r, &mdm.CommandResults{CommandUUID: "c1", Status: "Acknowledged", Raw: []byte(deviceInfoReport)})
	if err != nil {
		t.Fatal(err)
	}

	w := get("/v1/report/dev1/c1")
	if w.Code != http.StatusOK {
		t.Fatalf("got %d: %s", w.Code, w.Body)
	}
	var out httpapi.CommandReport
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != "Acknowledged" || out.Error != "" {
		t.Fatalf("status %q error %q", out.Status, out.Error)
	}
	q, _ := out.Report["QueryResponses"].(map[string]interface{})
	if q["iTunesStoreAccountIsActive"] != true || q["iTunesStoreAccountHash"] != "ctqKBqg9" {
		t.Fatalf("query responses not decoded: %#v", out.Report)
	}

	if w := get("/v1/report/dev1/other"); w.Code != http.StatusNotFound {
		t.Fatalf("unknown command: got %d, want 404", w.Code)
	}
	if w := get("/v1/report/dev1"); w.Code != http.StatusBadRequest {
		t.Fatalf("missing uuid: got %d, want 400", w.Code)
	}
}
