package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/micromdm/nanomdm/storage"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
	"github.com/micromdm/plist"
)

// CommandReport is the JSON body of the command report API.
type CommandReport struct {
	Status string `json:"status"`
	// Report is the device's raw report plist decoded into JSON. Data
	// values come out base64-encoded, dates as RFC 3339 strings.
	Report map[string]interface{} `json:"report,omitempty"`
	// Error is set when the raw report could not be decoded.
	Error string `json:"error,omitempty"`
}

// CommandReportHandler returns the report an enrollment sent for a
// command. The URL path is "<enrollment id>/<command uuid>", so the
// prefix needs to be stripped before this handler. It answers 404 with a
// JSON body while the enrollment has not reported on the command.
func CommandReportHandler(store storage.CommandReportRetriever, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		logger := ctxlog.Logger(r.Context(), logger)
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			http.Error(w, "path must be <enrollment id>/<command uuid>", http.StatusBadRequest)
			return
		}
		status, raw, found, err := store.RetrieveCommandReport(r.Context(), parts[0], parts[1])
		if err != nil {
			logger.Info("msg", "retrieving command report", "id", parts[0], "command_uuid", parts[1], "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "application/json")
		if !found {
			// JSON, unlike the mux's own 404, so a client can tell "not
			// reported yet" from "this server has no report API".
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(CommandReport{Error: "no report"})
			return
		}
		out := CommandReport{Status: status}
		if len(raw) > 0 {
			if err := plist.Unmarshal(raw, &out.Report); err != nil {
				out.Error = "decoding report: " + err.Error()
			}
		}
		if err := json.NewEncoder(w).Encode(out); err != nil {
			logger.Info("msg", "encoding json", "err", err)
		}
	}
}
