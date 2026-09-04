# Ultimate Tournament Results

Historic archive of ultimate tournament results, recorded using the Live! by BULA API.

[Live! by BULA](https://github.com/layoutd/live-by-bula) is the public results platform used by recent WFDF, EUC and BULA events. It serves an API to provide dynamic data for the frontend as cached JSON files. However, this data is not guaranteed to be hosted forever.

This project aims to provide a long term archive of tournament statistics for the general use of ultimate data enthusiasts.

## Workflow

1. **Archive & Publish** a tournament once it's finished. Find its host, base path and spec version, and run `utr` with that version (see [compatibility](#compatibility))

   ```
   dist/utr -version v01_09_14 -host wbuc.wfdf.sport
   ```

   This writes `data/<slug>.db`, where `<slug>` defaults to the season id the deployment's own heartbeat reports. This will error if the .db file already exists.

   It then Publishes the archive to `docs/archive/<slug>/` to host static JSON under Live! APIs original filenames. This is then a drop in replacement for existing tools.

2. **Add** the tournament to [`tournaments.csv`](tournaments.csv), then run `just fmt` to update readme and homepage.

Team photos and other media event links are not currently archived at all.

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
| `internal/liveversion/`   | registry mapping a -version flag value to that version's Archive/Publish entrypoints |
| `internal/convert`        | type conversion helpers                                                              |
| `cmd/utr/`                | the single CLI tool                                                                  |
| `cmd/gensite/`            | writes tournament info to readme and homepage from `tournaments.csv`                 |
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

## Archived tournaments

Generated from [`tournaments.csv`](tournaments.csv) by `just index`
do not edit this table by hand, edit the csv and regenerate instead

<!-- tournaments:start -->

| Event             | Date       | Host                                  | Version                       | [Legacy flags](#legacy-deployments) | Archive                                | Notes |
| ----------------- | ---------- | ------------------------------------- | ----------------------------- | ----------------------------------- | -------------------------------------- | ----- |
| EBUCC 2023        | 2023-06-09 | `live.ebucc.eu`                       | `v1.9.14` (`20250612.082841`) | `-unprefixed`                       | [`EBUCC2023`](docs/archive/EBUCC2023/) |       |
| WBUCC 2024        | 2024-10-14 | `live.wbucc.org`                      | `v1.9.14`                     | `-season-id` `-unprefixed`          | [`wbucc2024`](docs/archive/wbucc2024/) |       |
| EBUCC 2025        | 2025-06-06 | `live.ebucc.eu`                       | `v1.9.14` (`20250612.082841`) | `-unprefixed`                       | [`ebucc2025`](docs/archive/ebucc2025/) |       |
| WWUC 2025         | 2025-09-18 | `results.wfdf.sport/wwuc`             | `v1.9.14` (`dev`)             |                                     | [`WWUC2025`](docs/archive/WWUC2025/)   |       |
| EUCF 2025         | 2025-09-26 | `eucf.ultimatefederation.eu`          | `v1.9.14` (`v1.8.2`)          | `-season-id`                        | [`e2cf25`](docs/archive/e2cf25/)       |       |
| WBUC 2025         | 2025-11-16 | `wbuc.wfdf.sport`                     | `v1.9.14`                     |                                     | [`wbuc2025`](docs/archive/wbuc2025/)   |       |
| PAUC 2025         | 2025-12-01 | `results.pauc.sport`                  | `v1.9.14` (`v1.9.15`)         |                                     | [`pauc2025`](docs/archive/pauc2025/)   |       |
| EUIC 2026         | 2026-01-29 | `euic-schedule.ultimatefederation.eu` | `v1.9.14` (`v1.9.16`)         |                                     | [`euic2026`](docs/archive/euic2026/)   |       |
| Elite Invite 2026 | 2026-05-23 | `elite-invite.ultimatefederation.eu`  | `v1.9.14` (`v1.9.17`)         |                                     | [`26ELITLEU`](docs/archive/26ELITLEU/) |       |
| WMUCC 2026        | 2026-06-28 | `wmucc.wfdf.sport`                    | `v1.9.14` (`v1.9.17`)         |                                     | [`wmucc2026`](docs/archive/wmucc2026/) |       |
| WJUC 2026         | 2026-07-11 | `wjuc.wfdf.sport`                     | `v1.9.14` (`v1.9.17`)         |                                     | [`wjuc2026`](docs/archive/wjuc2026/)   |       |
| EYUC U17 2026     | 2026-08-03 | `eyuc-schedule.ultimatefederation.eu` | `v1.9.14` (`v1.9.17`)         |                                     | [`26EYUCVIE`](docs/archive/26EYUCVIE/) |       |
| WUCC 2026         | 2026-08-15 | `results.wfdf.sport/wucc-2026`        | `v3.0.6`                      |                                     | [`WUCC2026`](docs/archive/WUCC2026/)   |       |

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
| [Go](https://go.dev/dl/) 1.26+                                     | building/running `cmd/utr`, `cmd/gensite`, `gofmt`                                 |
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
