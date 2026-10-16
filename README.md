# Ultimate Tournament Results

Historic archive of ultimate tournament results, recorded using the Live! by BULA API, or for events without a Live! deployment, directly from their UltiOrganizer site.

[Live! by BULA](https://github.com/layoutd/live-by-bula) is the public results platform used by recent WFDF, EUC and BULA events. It serves an API to provide dynamic data for the frontend as cached JSON files. However, this data is not guaranteed to be hosted forever.

This project aims to provide a long term archive of tournament statistics for the general use of ultimate data enthusiasts.

## Workflow

1. **Archive & Publish** a tournament once it's finished. Find its host, base path and spec version, and run `utr` with that version (see [compatibility](#compatibility))

   ```
   dist/utr -version v01_09_14 -host wbuc.wfdf.sport
   ```

   This writes `data/<slug>.db`, where `<slug>` defaults to the season id the deployment's own heartbeat reports. This will error if the .db file already exists.

   It then Publishes the archive to `docs/archive/<slug>/` to host static JSON under Live! APIs original filenames. This is then a drop in replacement for existing tools.

   For an event run on plain UltiOrganizer with no Live! deployment, use `-version uo_01` with the UltiOrganizer install's own path as `-base-path` (see [UltiOrganizer sites](#ultiorganizer-sites))

   ```
   dist/utr -version uo_01 -host results.wfdf.sport -base-path /wmuc/
   ```

   A site that hosts several events, one season each (e.g. `eucs-schedule.ultimatefederation.eu` for the EUF Spring Tour), needs `-season-id` to pick one; the ids are in the site's `?view=seasonlist`

   ```
   dist/utr -version uo_01 -host eucs-schedule.ultimatefederation.eu -base-path / -season-id 26SPRIBIR
   ```

   A Live! archive can then be extended with the UltiOrganizer site underneath it, which turns it into a `uo_01` archive (see [Live! archives](#live-archives)). The site defaults to the archive's own host and base path less `live/data/`

   ```
   dist/utr -version v01_09_14 -mode enrich -db data/wmucc2026.db -host results.wfdf.sport -base-path /wmucc-2026/
   ```

2. **Add** the tournament to [`tournaments.csv`](tournaments.csv), then run `just fmt` to update readme and homepage. Set `live` to `y` while the original deployment still serves the tournament so the host links to it, and blank it once it's gone. `organiser` is one or more of `WFDF`, `EUF`, `UKU`, `BULA` separated by spaces with `BULA` always first (e.g. `BULA EUF`), add a new one to `organisers` in [`cmd/updatetables/main.go`](cmd/updatetables/main.go).

Team photos and other media event links are not archived from Live! deployments. `uo_01` keeps the links (not the media itself) shown on UltiOrganizer team and game pages.

## How this works

The Tournaments API is polled (with rate limited) to gather all data from the tournament. This is then archived into a SQLite file with a schema very similar to that of UltiOrganizer. Some edits are made due to aggregate values being exposed so the underlying data cannot be recreated.

### Versioning

As there are many versions of Live! deployed, the versioning approach decided on in this project was to have a single produced binary (`utr`) with many version flags. underlying GO code is kept as generic as possible and then split into versioned sub-packages where required.

Package names are the Live! app version with each dot-separated segment zero-padded to two digits and joined by underscores (`1.9.14` -> `v01_09_14`).

Where multiple Live! versions have very similar response and schema shapes they are covered by the same version flag. The degree to which versions are considered similar is achived by diffing the source code and responses from deployments then judging if there is a significant different. Therefore a version key can cover a range of releases (i.e. 1.9.14 -> 1.9.17).

### Project structure

| Package                   | Description                                                                          |
| ------------------------- | ------------------------------------------------------------------------------------ |
| `db/`                     | SQL schema and query for sqlc                                                        |
| `internal/db/`            | sqlc-generated db interface package                                                  |
| `internal/livedatamodel/` | Live! API's response shapes                                                          |
| `internal/liveclient/`    | function to gather information from Live! deployment                                 |
| `internal/livearchive/`   | archive workflow (gather data, apply schema, write sqlite)                           |
| `internal/livepublish/`   | render sqlite db as static JSON in docs/archive/                                     |
| `internal/store/`         | plain-Go-typed reads and writes over each version's sqlc package                     |
| `internal/uoclient/`      | scrape one season from a classic UltiOrganizer site's HTML pages and CSV exports     |
| `internal/uoarchive/`     | resolve a scraped UltiOrganizer season into uo_01 rows, for a new or a Live! archive |
| `internal/liveversion/`   | registry mapping a -version flag value to that version's Archive/Publish entrypoints |
| `internal/convert`        | type conversion helpers                                                              |
| `cmd/utr/`                | the single CLI tool                                                                  |
| `cmd/updatetables/`       | writes tournament info to readme and homepage from `tournaments.csv`                 |
| `data/`                   | data archive, one sqlite .db file per tournament                                     |
| `docs/`                   | GitHub Pages web root                                                                |
| `docs/index.html`         | homepage                                                                             |
| `docs/archive/`           | published static JSON for archived tournaments                                       |

## Compatibility

| Live! by BULA app version | `-version` flag | Status      |
| ------------------------- | --------------- | ----------- |
| 1.9.14 - 1.9.17           | v1.9.14         | Published   |
| 2.0.0                     | -               | Not Planned |
| 3.0.6                     | v3.0.6          | Published   |

| Source                                   | `-version` flag | Status    |
| ---------------------------------------- | --------------- | --------- |
| UltiOrganizer site with no Live! install | uo_01           | Published |

### UltiOrganizer sites

Many events, e.g. about half of those listed on [results.wfdf.sport](https://results.wfdf.sport), ran on UltiOrganizer without Live! on top. `uo_01` archives those by scraping the site's public pages (schedule, pools, team cards, game pages) and CSV exports, since UltiOrganizer's own `ext/restful.php` API is broken on current deployments.

The archive is a superset of the `v1.9.14` schema: [`db/v01_09_14/schema.sql`](db/v01_09_14/schema.sql) unchanged, plus [`db/uo_01/schema.sql`](db/uo_01/schema.sql) for what UltiOrganizer shows publicly that the Live! API has no place for. So it publishes the same Live! JSON as any 1.9.14-1.9.17 archive, and the extras are kept in `data/<slug>.db` only:

| Table                 | Holds                                                                     |
| --------------------- | ------------------------------------------------------------------------- |
| `played`              | everyone entered as playing each game, with their number and captain flag |
| `game_events`         | timeouts, turnovers and starting offence from each scoresheet             |
| `team_pools`          | every pool each team was in, with its round-robin rank                    |
| `game_pools`          | every pool each game counts towards                                       |
| `move_teams`          | which round-robin placing continues to which pool                         |
| `clubs`, `team_clubs` | team clubs, by id                                                         |
| `team_profiles`       | coach, captain, story and achievements from team cards                    |
| `urls`                | team pages and photo/video links on team and game pages                   |
| `server_settings`     | server settings the public pages follow, e.g. whether to show comments    |

Locations and reservations also keep the address, coordinates and times the Live! API leaves out.

Where UltiOrganizer shows something only indirectly, the archive notes how it was derived, in [`db/uo_01/schema.sql`](db/uo_01/schema.sql) and [`internal/uoarchive`](internal/uoarchive/):

- the season's start/end are its first and last reservation, and its type isn't shown at all
- pool type is read from how the pool page renders, and a pool's rules (winning score, caps, timeouts) aren't shown
- home/visitor slot names (e.g. "Winner QF1") have no id anywhere, so they get synthetic negative ids
- spirit comes from the `ext/spiritcsv.php` export when the site serves it: every submitted score with its comments, including games where only one team has submitted (the game page hides those until both have). Where a site locks the export, it falls back to the game pages, which only show two-sided scores and show comments only if the season enables it
- a game with no scoresheet (nobody took stats, so it has just its result and any spirit, or it was a forfeit or never played) isn't linked, so has no id anywhere, and gets a synthetic negative id in schedule order. It has no goals, played or events. A whole division scored that way (e.g. WUC 2024's Guts) can still be left out with `-skip-divisions`
- where team cards error (every team on some 2023 sites), the roster comes from the team's scoreboard (`view=scorestatus`), which runs the same roster query, with shirt numbers from the players CSV. Those teams have no profile, club id or team links
- a round-robin pool with exactly one game redirects to that game's page, so its name comes from the left menu and it has no standings
- publish only includes spirit comments when the server showed them, read from whether game pages have a comments row

### Live! archives

`-mode enrich` adds the same extras to an existing 1.9.14-1.9.17 Live! archive from the UltiOrganizer site underneath it, making it a `uo_01` archive. The Live! archive stays the base: it has every game, including ones UltiOrganizer has no id for. Everything the two share (games, teams, players, goals, spirit) is checked to agree first, and any difference stops it before anything is written. It runs once per archive.

Its published JSON stays as Live! served it, apart from spirit scores only one team submitted, which UltiOrganizer has and the Live! API leaves out.

## Archived tournaments

Generated from [`tournaments.csv`](tournaments.csv) by `just index`
do not edit this table by hand, edit the csv and regenerate instead

<!-- tournaments:start -->

| Event                        | Organiser  | Date       | Host                                                                                    | Version                       | [Legacy flags](#legacy-deployments) | Archive                                | Notes                        |
| ---------------------------- | ---------- | ---------- | --------------------------------------------------------------------------------------- | ----------------------------- | ----------------------------------- | -------------------------------------- | ---------------------------- |
| EBUCC 2023                   | BULA, EUF  | 2023-06-09 | [`live.ebucc.eu/scores2023`](https://live.ebucc.eu/scores2023/)                         | `v1.9.14` (`20250612.082841`) | `-unprefixed`                       | [`EBUCC2023`](docs/archive/EBUCC2023/) |                              |
| WU24 2023                    | WFDF       | 2023-07-02 | [`results.wfdf.sport/wu24-2023`](https://results.wfdf.sport/wu24-2023/)                 | `uo_01`                       |                                     | [`WU24-2023`](docs/archive/WU24-2023/) |                              |
| EUCR-S Open Bern 2023        | EUF        | 2023-08-26 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`EUCR23SUI`](docs/archive/EUCR23SUI/) | No stats taken               |
| EUCF 2023                    | EUF        | 2023-09-29 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`EUCF23`](docs/archive/EUCF23/)       |                              |
| WBUC 2023                    | BULA, WFDF | 2023-11-01 | [`results.wfdf.sport/wbuc-2023`](https://results.wfdf.sport/wbuc-2023/)                 | `uo_01`                       |                                     | [`WBUC2023`](docs/archive/WBUC2023/)   |                              |
| PAUC 2023                    | WFDF       | 2023-11-28 | [`results.wfdf.sport/pauc-2023`](https://results.wfdf.sport/pauc-2023/)                 | `uo_01`                       |                                     | [`PAUC2023`](docs/archive/PAUC2023/)   |                              |
| AOBUC 2024                   | WFDF       | 2024-06-12 | [`results.wfdf.sport/aobuc`](https://results.wfdf.sport/aobuc/)                         | `uo_01`                       |                                     | [`AOBUC24`](docs/archive/AOBUC24/)     |                              |
| WJUC 2024                    | WFDF       | 2024-07-21 | [`results.wfdf.sport/wjuc-2024`](https://results.wfdf.sport/wjuc-2024/)                 | `uo_01`                       |                                     | [`WJUC2024`](docs/archive/WJUC2024/)   |                              |
| WUC 2024                     | WFDF       | 2024-08-31 | [`results.wfdf.sport/wuc`](https://results.wfdf.sport/wuc/)                             | `uo_01`                       |                                     | [`WUC2024`](docs/archive/WUC2024/)     | Guts divisions left out      |
| EUCF 2024                    | EUF        | 2024-09-27 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`eucf24`](docs/archive/eucf24/)       | No stats for 61 of 156 games |
| WBUCC 2024                   | BULA       | 2024-10-14 | [`live.wbucc.org`](https://live.wbucc.org/)                                             | `v1.9.14`                     | `-season-id` `-unprefixed`          | [`wbucc2024`](docs/archive/wbucc2024/) |                              |
| WMUC 2024                    | WFDF       | 2024-11-10 | [`results.wfdf.sport/wmuc`](https://results.wfdf.sport/wmuc/)                           | `uo_01`                       |                                     | [`WMUC2024`](docs/archive/WMUC2024/)   |                              |
| Spring Tour Timisoara 2025   | EUF        | 2025-04-26 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`st-timi25`](docs/archive/st-timi25/) |                              |
| Spring Tour Lund 2025        | EUF        | 2025-04-26 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`st-lund25`](docs/archive/st-lund25/) |                              |
| Spring Tour Leicester 2025   | EUF        | 2025-04-26 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`st-leic25`](docs/archive/st-leic25/) |                              |
| Spring Tour Grenoble 2025    | EUF        | 2025-04-26 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`st-gren25`](docs/archive/st-gren25/) |                              |
| Spring Tour Padova 2025      | EUF        | 2025-04-26 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`padova25`](docs/archive/padova25/)   |                              |
| Elite Invite 2025            | EUF        | 2025-05-30 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`EI_25`](docs/archive/EI_25/)         |                              |
| Spring Tour Espoo 2025       | EUF        | 2025-05-31 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`st-espo25`](docs/archive/st-espo25/) |                              |
| EBUCC 2025                   | BULA, EUF  | 2025-06-06 | [`live.ebucc.eu/scores2025`](https://live.ebucc.eu/scores2025/)                         | `v1.9.14` (`20250612.082841`) | `-unprefixed`                       | [`ebucc2025`](docs/archive/ebucc2025/) |                              |
| WU24 2025                    | WFDF       | 2025-06-21 | [`results.wfdf.sport/wu24`](https://results.wfdf.sport/wu24/)                           | `uo_01`                       |                                     | [`wu24-2025`](docs/archive/wu24-2025/) |                              |
| Summer Tour Warsaw 2025      | EUF        | 2025-08-02 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`sumt-wars`](docs/archive/sumt-wars/) |                              |
| Summer Tour Hornsyld 2025    | EUF        | 2025-08-02 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`sumt-horn`](docs/archive/sumt-horn/) |                              |
| Summer Tour Stirling 2025    | EUF        | 2025-08-02 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`sumt-stir`](docs/archive/sumt-stir/) | No stats taken               |
| Summer Tour Heilbronn 2025   | EUF        | 2025-08-02 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`sumt-heil`](docs/archive/sumt-heil/) |                              |
| Summer Tour Wroclaw 2025     | EUF        | 2025-08-16 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`sumt-wroc`](docs/archive/sumt-wroc/) |                              |
| Summer Tour Rimini 2025      | EUF        | 2025-08-16 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`sumt-rimi`](docs/archive/sumt-rimi/) |                              |
| AOUC 2025                    | WFDF       | 2025-09-03 | [`results.wfdf.sport/aouc`](https://results.wfdf.sport/aouc/)                           | `v1.9.14` (`v1.8.3`)          | `-season-id`                        | [`AOUC2025`](docs/archive/AOUC2025/)   |                              |
| WWUC 2025                    | WFDF       | 2025-09-18 | [`results.wfdf.sport/wwuc`](https://results.wfdf.sport/wwuc/)                           | `v1.9.14` (`dev`)             |                                     | [`WWUC2025`](docs/archive/WWUC2025/)   |                              |
| E2CF 2025                    | EUF        | 2025-09-20 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`eucf25`](docs/archive/eucf25/)       |                              |
| EUCF 2025                    | EUF        | 2025-09-26 | `eucf.ultimatefederation.eu`                                                            | `v1.9.14` (`v1.8.2`)          | `-season-id`                        | [`e2cf25`](docs/archive/e2cf25/)       |                              |
| AAUC 2025                    | WFDF       | 2025-10-09 | [`results.wfdf.sport/aauc`](https://results.wfdf.sport/aauc/)                           | `v1.9.14` (`dev`)             | `-season-id`                        | [`AAUC2025`](docs/archive/AAUC2025/)   |                              |
| WBUC 2025                    | BULA, WFDF | 2025-11-16 | [`results.wfdf.sport/wbuc-2025`](https://results.wfdf.sport/wbuc-2025/)                 | `v1.9.14`                     |                                     | [`wbuc2025`](docs/archive/wbuc2025/)   |                              |
| PAUC 2025                    | WFDF       | 2025-12-01 | [`results.wfdf.sport/pauc-2025`](https://results.wfdf.sport/pauc-2025/)                 | `v1.9.14` (`v1.9.15`)         |                                     | [`pauc2025`](docs/archive/pauc2025/)   |                              |
| EUIC 2026                    | EUF        | 2026-01-29 | [`euic-schedule.ultimatefederation.eu`](https://euic-schedule.ultimatefederation.eu/)   | `v1.9.14` (`v1.9.16`)         |                                     | [`euic2026`](docs/archive/euic2026/)   |                              |
| Spring Tour Padova 2026      | EUF        | 2026-04-18 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SPRIPAD`](docs/archive/26SPRIPAD/) |                              |
| Spring Tour Budweis 2026     | EUF        | 2026-04-18 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SPRIBUD`](docs/archive/26SPRIBUD/) |                              |
| Spring Tour Angers 2026      | EUF        | 2026-04-18 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SPRIANG`](docs/archive/26SPRIANG/) |                              |
| Spring Tour Birmingham 2026  | EUF        | 2026-04-25 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SPRIBIR`](docs/archive/26SPRIBIR/) |                              |
| Spring Tour Grenoble 2026    | EUF        | 2026-04-25 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SPRIGRE`](docs/archive/26SPRIGRE/) |                              |
| Spring Tour Hornsyld 2026    | EUF        | 2026-04-25 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SPRIHOR`](docs/archive/26SPRIHOR/) |                              |
| EMUCC 2026                   | EUF        | 2026-05-08 | [`emucc-schedule.ultimatefederation.eu`](https://emucc-schedule.ultimatefederation.eu/) | `v1.9.14` (`v1.9.17`)         |                                     | [`26EMUCC`](docs/archive/26EMUCC/)     |                              |
| Elite Invite 2026            | EUF        | 2026-05-23 | [`elite-invite.ultimatefederation.eu`](https://elite-invite.ultimatefederation.eu/)     | `v1.9.14` (`v1.9.17`)         |                                     | [`26ELITLEU`](docs/archive/26ELITLEU/) |                              |
| Bad Skid Invite 2026         | EUF        | 2026-06-20 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26BADSKID`](docs/archive/26BADSKID/) |                              |
| WMUCC 2026                   | WFDF       | 2026-06-28 | [`results.wfdf.sport/wmucc-2026`](https://results.wfdf.sport/wmucc-2026/)               | `v1.9.14` (`v1.9.17`)         |                                     | [`wmucc2026`](docs/archive/wmucc2026/) |                              |
| WJUC 2026                    | WFDF       | 2026-07-11 | [`results.wfdf.sport/wjuc-2026`](https://results.wfdf.sport/wjuc-2026/)                 | `v1.9.14` (`v1.9.17`)         |                                     | [`wjuc2026`](docs/archive/wjuc2026/)   |                              |
| Summer Tour Copenhagen 2026  | EUF        | 2026-07-11 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SUMCOPE`](docs/archive/26SUMCOPE/) |                              |
| Summer Tour Wroclaw 2026     | EUF        | 2026-07-25 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SUMWROC`](docs/archive/26SUMWROC/) |                              |
| Summer Tour Southampton 2026 | EUF        | 2026-07-25 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SUMSOUT`](docs/archive/26SUMSOUT/) |                              |
| Summer Tour Bordeaux 2026    | EUF        | 2026-07-25 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SUMBORD`](docs/archive/26SUMBORD/) | No stats taken               |
| Summer Tour Warsaw 2026      | EUF        | 2026-07-25 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SUMWARS`](docs/archive/26SUMWARS/) |                              |
| Summer Tour Rimini 2026      | EUF        | 2026-07-25 | [`eucs-schedule.ultimatefederation.eu`](https://eucs-schedule.ultimatefederation.eu/)   | `uo_01`                       |                                     | [`26SUMRIMI`](docs/archive/26SUMRIMI/) |                              |
| EYUC U17 2026                | EUF        | 2026-08-03 | [`eyuc-schedule.ultimatefederation.eu`](https://eyuc-schedule.ultimatefederation.eu/)   | `v1.9.14` (`v1.9.17`)         |                                     | [`26EYUCVIE`](docs/archive/26EYUCVIE/) |                              |
| WUCC 2026                    | WFDF       | 2026-08-15 | [`results.wfdf.sport/wucc-2026`](https://results.wfdf.sport/wucc-2026/)                 | `v3.0.6`                      |                                     | [`WUCC2026`](docs/archive/WUCC2026/)   |                              |
| UK Nats 2026                 | UKU        | 2026-09-05 | [`uku-schedule.ultimatefederation.eu`](https://uku-schedule.ultimatefederation.eu/)     | `v1.7.8` (`v1.9.14`)          |                                     | [`26UKUNATS`](docs/archive/26UKUNATS/) |                              |
| EUCF 2026                    | EUF        | 2026-10-09 | [`eucf.ultimatefederation.eu`](https://eucf.ultimatefederation.eu/)                     | `v1.9.14` (`v1.9.17`)         |                                     | [`26EUCFWRO`](docs/archive/26EUCFWRO/) |                              |

<!-- tournaments:end -->

### Legacy deployments

A handful of known deployments serve the exact 1.9.14-1.9.16 response shape but don't follow its normal conventions -- see live-by-bula-openapi's own [Legacy and unsupported deployments](https://github.com/cxd309/live-by-bula-openapi#legacy-and-unsupported-deployments) section for the full list and why each one is flagged. `utr -version v01_09_14` covers them with two extra flags, both additive: the archive and its published output end up looking exactly like a normal 1.9.14-1.9.17 tournament either way.

- `-season-id` overrides the season id instead of discovering it from the heartbeat, for deployments whose heartbeat has no `config` block, or no heartbeat endpoint at all
- `-unprefixed` for deployments that serve static filenames with no season-id prefix (`reference.json` instead of `{seasonId}_reference.json`)

```
dist/utr -version v01_09_14 -host live.wbucc.org -unprefixed -season-id wbucc2024
```

## Dependencies

Everything below is invoked from the [`justfile`](justfile), run `just check-deps`
to confirm they're all on `PATH`

| Tool                                                               | Used for                                                                           |
| ------------------------------------------------------------------ | ---------------------------------------------------------------------------------- |
| [Go](https://go.dev/dl/) 1.26+                                     | building/running `cmd/utr`, `cmd/updatetables`, `gofmt`                            |
| [just](https://github.com/casey/just)                              | running the recipes below                                                          |
| [golangci-lint](https://golangci-lint.run/)                        | `just fmt`                                                                         |
| [dprint](https://dprint.dev/)                                      | `just fmt`, formats markdown/json/toml/yaml                                        |
| [pgFormatter](https://github.com/darold/pgFormatter) (`pg_format`) | `just fmt`, formats `db/**/*.sql`                                                  |
| [sqlc](https://sqlc.dev/)                                          | `just dbgen`                                                                       |
| [simple-file-server](https://github.com/cxd309/simple-file-server) | `just serve`, install via `go install github.com/cxd309/simple-file-server@latest` |

## Development usage

```
just check-deps   # confirm required CLI tools are installed
just index        # regenerate README table + docs/tournaments.csv from tournaments.csv
just dbgen        # sqlc generate internal/db/ packages from db/ with
just build        # compile the utr binary into dist/
just serve        # serve docs/ locally, emulating GitHub Pages
just clean        # remove dist/
just fmt          # go mod tidy, gofmt, golangci-lint, dprint, pg_format
```
