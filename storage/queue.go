package storage

import (
	"context"

	"github.com/micromdm/nanomdm/mdm"
)

// CommandAndReportResultsStore stores and retrieves MDM command queue data.
type CommandAndReportResultsStore interface {
	StoreCommandReport(r *mdm.Request, report *mdm.CommandResults) error
	RetrieveNextCommand(r *mdm.Request, skipNotNow bool) (*mdm.Command, error)
	ClearQueue(r *mdm.Request) error
}

// CommandEnqueuer is able to enqueue MDM commands.
type CommandEnqueuer interface {
	EnqueueCommand(ctx context.Context, id []string, cmd *mdm.Command) (map[string]error, error)
}

// CommandReportRetriever retrieves the report an enrollment sent for a command.
type CommandReportRetriever interface {
	// RetrieveCommandReport returns the status and raw plist of the last
	// report enrollment id sent for commandUUID. found is false when the
	// enrollment has not reported on that command (yet).
	RetrieveCommandReport(ctx context.Context, id, commandUUID string) (status string, raw []byte, found bool, err error)
}
