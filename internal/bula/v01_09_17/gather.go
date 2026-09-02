package v01_09_17

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
	GameDetailByID map[int64]*GameDetailResponse
}

// Gather fetches every response this archiver needs from one deployment: the heartbeat,
// the reference endpoint, every team's detail, the games list (to enumerate game ids),
// and every game's detail -- one request per team and per game.
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

	games, err := client.FetchGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch games: %w", err)
	}

	detailByGameID := make(map[int64]*GameDetailResponse, len(games.Games))
	for _, game := range games.Games {
		detail, err := client.FetchGameDetail(ctx, game.GameID)
		if err != nil {
			return nil, fmt.Errorf("fetch game detail %d: %w", game.GameID, err)
		}
		detailByGameID[detail.GameResult.GameID] = detail
	}

	return &Snapshot{
		Heartbeat:      hb,
		Reference:      ref,
		TeamDetailByID: detailByTeamID,
		GameDetailByID: detailByGameID,
	}, nil
}
