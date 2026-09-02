package liveclient

import (
	"context"
	"fmt"

	"github.com/cxd309/ultimate-tournament-results/internal/livedatamodel/v03_00_06"
)

// Snapshot bundles everything Gather fetches from one deployment in a single archive
// run, so livearchive's Import can write it all in one pass
type Snapshot struct {
	Heartbeat         *livedatamodel.HeartbeatResponse
	Reference         *livedatamodel.ReferenceResponse
	TeamDetailByID    map[int64]*livedatamodel.TeamDetailResponse
	GameDetailByID    map[int64]*livedatamodel.GameDetailResponse
	GamePoolsByGameID map[int64][]int64
}

// Gather fetches every response one archive run needs from this deployment: the
// heartbeat, the reference endpoint, every team's detail, the games list (to enumerate
// game ids), and every game's detail -- one request per team and per game
func (c *Client) Gather(ctx context.Context) (*Snapshot, error) {
	hb, err := c.FetchHeartbeat(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch heartbeat: %w", err)
	}

	ref, err := c.FetchReference(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch reference: %w", err)
	}

	detailByTeamID := make(map[int64]*livedatamodel.TeamDetailResponse, len(ref.Teams))
	for _, team := range ref.Teams {
		detail, err := c.FetchTeamDetail(ctx, team.TeamID)
		if err != nil {
			return nil, fmt.Errorf("fetch team detail %d: %w", team.TeamID, err)
		}
		detailByTeamID[detail.TeamID] = detail
	}

	games, err := c.FetchGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch games: %w", err)
	}

	detailByGameID := make(map[int64]*livedatamodel.GameDetailResponse, len(games.Games))
	poolsByGameID := make(map[int64][]int64, len(games.Games))
	for _, game := range games.Games {
		detail, err := c.FetchGameDetail(ctx, game.GameID)
		if err != nil {
			return nil, fmt.Errorf("fetch game detail %d: %w", game.GameID, err)
		}
		detailByGameID[detail.GameResult.GameID] = detail
		poolsByGameID[game.GameID] = game.Pools
	}

	return &Snapshot{
		Heartbeat:         hb,
		Reference:         ref,
		TeamDetailByID:    detailByTeamID,
		GameDetailByID:    detailByGameID,
		GamePoolsByGameID: poolsByGameID,
	}, nil
}
