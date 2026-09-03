package liveclient

import (
	"context"
	"fmt"

	"github.com/cxd309/ultimate-tournament-results/internal/liveclient"
	"github.com/cxd309/ultimate-tournament-results/internal/livedatamodel/v01_09_14"
)

// Snapshot bundles everything Gather fetches from one deployment in a single archive
// run, so livearchive's Import can write it all in one pass
type Snapshot struct {
	Heartbeat      *livedatamodel.HeartbeatResponse
	Reference      *livedatamodel.ReferenceResponse
	TeamDetailByID map[int64]*livedatamodel.TeamDetailResponse
	GameDetailByID map[int64]*livedatamodel.GameDetailResponse
}

// Gather fetches every response one archive run needs from this deployment: the
// heartbeat, the reference endpoint, every team's detail, the games list (to enumerate
// game ids), and every game's detail -- one request per team and per game
//
// prints its own progress: team/game detail is one HTTP request each, throttled,
// so a big tournament genuinely takes a while with nothing else to show for it
func (c *Client) Gather(ctx context.Context) (*Snapshot, error) {
	fmt.Println("fetching heartbeat...")
	hb, err := c.FetchHeartbeat(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch heartbeat: %w", err)
	}

	fmt.Println("fetching reference...")
	ref, err := c.FetchReference(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch reference: %w", err)
	}

	fmt.Printf("fetching %d team details...\n", len(ref.Teams))
	detailByTeamID := make(map[int64]*livedatamodel.TeamDetailResponse, len(ref.Teams))
	for i, team := range ref.Teams {
		detail, err := c.FetchTeamDetail(ctx, team.TeamID)
		if err != nil {
			return nil, fmt.Errorf("fetch team detail %d: %w", team.TeamID, err)
		}
		detailByTeamID[detail.TeamID] = detail
		liveclient.PrintProgress(i+1, len(ref.Teams), 10)
	}

	fmt.Println("fetching games list...")
	games, err := c.FetchGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch games: %w", err)
	}

	fmt.Printf("fetching %d game details...\n", len(games.Games))
	detailByGameID := make(map[int64]*livedatamodel.GameDetailResponse, len(games.Games))
	for i, game := range games.Games {
		detail, err := c.FetchGameDetail(ctx, game.GameID)
		if err != nil {
			return nil, fmt.Errorf("fetch game detail %d: %w", game.GameID, err)
		}
		detailByGameID[detail.GameResult.GameID] = detail
		liveclient.PrintProgress(i+1, len(games.Games), 25)
	}

	return &Snapshot{
		Heartbeat:      hb,
		Reference:      ref,
		TeamDetailByID: detailByTeamID,
		GameDetailByID: detailByGameID,
	}, nil
}
