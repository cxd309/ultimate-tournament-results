package v1914

import (
	"context"
	"fmt"
)

// Snapshot bundles everything Gather fetches from one deployment in a single archive
// run, so Import can write it all in one pass.
type Snapshot struct {
	Heartbeat      *HeartbeatResponse
	Reference      *ReferenceResponse
	TeamDetailByID map[int64]*TeamDetailResponse
}

// Gather fetches every response this archiver needs from one deployment: the heartbeat,
// the reference endpoint, and every team's detail (one request per team).
func Gather(ctx context.Context, client *Client) (*Snapshot, error) {
	hb, err := client.FetchHeartbeat(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch heartbeat: %w", err)
	}

	ref, err := client.FetchReference(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch reference: %w", err)
	}

	detailByTeamID := make(map[int64]*TeamDetailResponse, len(ref.Teams))
	for _, team := range ref.Teams {
		detail, err := client.FetchTeamDetail(ctx, team.TeamID)
		if err != nil {
			return nil, fmt.Errorf("fetch team detail %d: %w", team.TeamID, err)
		}
		detailByTeamID[detail.TeamID] = detail
	}

	return &Snapshot{
		Heartbeat:      hb,
		Reference:      ref,
		TeamDetailByID: detailByTeamID,
	}, nil
}
