# ultimate-tournament-results

Historic archive of ultimate tournament results, recorded using the Live! by BULA API.

Live! by BULA (API documented in the repo [live-by-bula-openapi](https://github.com/cxd309/live-by-bula-openapi)) is the public results platform used by recent WFDF, EUC and BULA events. It serves each tournament's data as JSON files for as long as the event stays publicly available, then the files disappear. This repo polls known deployments while they're still up and keeps the results permanently.

## How this works

Each tournament is archived into its own SQLite file, normalized to a schema matched exactly to whichever Live! spec version that deployment runs.

There is no attempt to force one schema to cover every API version, so each spec version gets its own schema, its own generated query code, and its own archive/publish implementation, all following the same pattern:

```
db/v01_09_14/                      SQL source: schema.sql + query.sql, read by sqlc
internal/db/v01_09_14/             sqlc-generated db interface package for GO
internal/store/v01_09_14/          plain-Go-typed wrapper over internal/db/v01_09_14
internal/livedatamodel/v01_09_14/  the Live! 1.9.14-1.9.17 API's response shapes
internal/liveclient/               shared HTTP transport every version's Client embeds
internal/liveclient/v01_09_14/     typed Fetch* methods + Gather, using livedatamodel
internal/livearchive/              shared archive workflow (gather, apply schema, import,
                                   commit, read back) every version's Archive plugs into
internal/livearchive/v01_09_14/    maps a gathered snapshot to the store, and wires
                                   Archive for the registry
internal/livepublish/v01_09_14/    renders a data/<slug>.db as static JSON in docs/archive/
internal/liveversion/              registry mapping a -version flag value to that
                                   version's Archive/Publish entrypoints
internal/convert/                  sql to plain go type conversions
cmd/utr/                           the single CLI: -version selects the spec
                                   implementation, -mode selects archive/publish/all
cmd/gensite/                       regenerates docs/tournaments.csv and the README
                                   table from tournaments.csv, run via `just index`
data/                              data archive, one sqlite .db file per tournament
docs/                              GitHub Pages web root
docs/index.html                    homepage, renders docs/tournaments.csv as a table
docs/tournaments.csv               generated copy of tournaments.csv, fetched by
                                   docs/index.html
docs/archive/                      published static JSON, one folder per tournament
```

Package names (and `-version` flag values) are the Live! app version with each dot-separated segment zero-padded to two digits and joined by underscores (`1.9.14` -> `v01_09_14`). A version key can cover a range of releases when they turn out to behave identically for everything this repo archives (confirmed by diffing the actual PHP source, not assumed) -- the key is always the lowest release in that range.

Sharing stops at whatever's genuinely mechanical (HTTP fetch, transaction/commit plumbing) -- `livearchive/vXX/import.go`, `livedatamodel/vXX`, and `db/vXX/schema.sql` stay independent per version, since that's exactly where the API/schema actually diverges between versions.

## Compatibility

| Live! by BULA app version | `-version` flag | Status         |
| ------------------------- | --------------- | -------------- |
| 1.9.14 - 1.9.17           | v01_09_14       | In Development |
| 2.0.0                     | -               | Not Planned    |
| 3.0.6                     | v03_00_06       | In Development |

## Archived tournaments

Generated from [`tournaments.csv`](tournaments.csv) by `just index`
do not edit this table by hand, edit the csv and regenerate instead

<!-- tournaments:start -->

| Event             | Date       | Host                                  | Live!     | [Legacy flags](#legacy-deployments) | Archive                                |
| ----------------- | ---------- | ------------------------------------- | --------- | ----------------------------------- | -------------------------------------- |
| EBUCC 2023        | 2023-06-09 | `live.ebucc.eu`                       | v01_09_14 | `-unprefixed`                       | [`EBUCC2023`](docs/archive/EBUCC2023/) |
| WBUCC 2024        | 2024-10-14 | `live.wbucc.org`                      | v01_09_14 | `-season-id` `-unprefixed`          | [`wbucc2024`](docs/archive/wbucc2024/) |
| EBUCC 2025        | 2025-06-06 | `live.ebucc.eu`                       | v01_09_14 | `-unprefixed`                       | [`ebucc2025`](docs/archive/ebucc2025/) |
| WWUC 2025         | 2025-09-18 | `results.wfdf.sport/wwuc`             | v01_09_14 |                                     | [`WWUC2025`](docs/archive/WWUC2025/)   |
| EUCF 2025         | 2025-09-26 | `eucf.ultimatefederation.eu`          | v01_09_14 | `-season-id`                        | [`e2cf25`](docs/archive/e2cf25/)       |
| WBUC 2025         | 2025-11-16 | `wbuc.wfdf.sport`                     | v01_09_14 |                                     | [`wbuc2025`](docs/archive/wbuc2025/)   |
| PAUC 2025         | 2025-12-01 | `results.pauc.sport`                  | v01_09_14 |                                     | [`pauc2025`](docs/archive/pauc2025/)   |
| EUIC 2026         | 2026-01-29 | `euic-schedule.ultimatefederation.eu` | v01_09_14 |                                     | [`euic2026`](docs/archive/euic2026/)   |
| Elite Invite 2026 | 2026-05-23 | `elite-invite.ultimatefederation.eu`  | v01_09_14 |                                     | [`26ELITLEU`](docs/archive/26ELITLEU/) |
| WMUCC 2026        | 2026-06-28 | `wmucc.wfdf.sport`                    | v01_09_14 |                                     | [`wmucc2026`](docs/archive/wmucc2026/) |
| WJUC 2026         | 2026-07-11 | `wjuc.wfdf.sport`                     | v01_09_14 |                                     | [`wjuc2026`](docs/archive/wjuc2026/)   |
| EYUC U17 2026     | 2026-08-03 | `eyuc-schedule.ultimatefederation.eu` | v01_09_14 |                                     | [`26EYUCVIE`](docs/archive/26EYUCVIE/) |
| WUCC 2026         | 2026-08-15 | `results.wfdf.sport/wucc-2026`        | v03_00_06 |                                     | [`WUCC2026`](docs/archive/WUCC2026/)   |

<!-- tournaments:end -->

## Workflow

1. **Archive** a tournament once it's finished. Find its host, base path and spec version, and run `utr` with that version (see [compatibility](#compatibility))

   ```
   dist/utr -version v01_09_14 -mode archive -host wbuc.wfdf.sport
   ```

   This writes `data/<slug>.db`, where `<slug>` defaults to the season id the deployment's own heartbeat reports.

   Archives are write-once: the command refuses to run if that file already exists.

2. **Publish** the archive to `docs/archive/<slug>/` for GitHub Pages, via `-mode publish` or `-mode all`. Renders every endpoint's JSON under the live API's own relative filenames, so pointing a tool's `host`/`basePath` at the published folder is a drop-in replacement for the original deployment.

3. **Add** the tournament to [`tournaments.csv`](tournaments.csv), then run `just index` to regenerate the README table above and `docs/tournaments.csv`, which [`docs/index.html`](docs/index.html) renders as the homepage's tournament list.

   `tournaments.csv` is hand-maintained rather than read straight out of `data/*.db`: the display columns (event name, legacy footnote) are editorial judgement calls, not archive data, so a hand-edited row can't drift out of sync with an automated one.

Team photos and other media event links are not currently archived at all, they are real uploaded image files served by the deployment, not JSON data, so capturing them needs a future change to fetch and store the actual assets rather than just their metadata.

### Legacy deployments

A handful of known deployments serve the exact 1.9.14-1.9.16 response shape but don't follow its normal conventions -- see live-by-bula-openapi's own [Legacy and unsupported deployments](https://github.com/cxd309/live-by-bula-openapi#legacy-and-unsupported-deployments) section for the full list and why each one is flagged. `utr -version v01_09_14` covers them with two extra flags, both additive: the archive and its published output end up looking exactly like a normal 1.9.14-1.9.17 tournament either way.

- `-season-id` overrides the season id instead of discovering it from the heartbeat, for deployments whose heartbeat has no `config` block, or no heartbeat endpoint at all
- `-unprefixed` for deployments that serve static filenames with no season-id prefix (`reference.json` instead of `{seasonId}_reference.json`)

```
dist/utr -version v01_09_14 -mode all -host live.wbucc.org -base-path /live/data/ -unprefixed -season-id wbucc2024
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
