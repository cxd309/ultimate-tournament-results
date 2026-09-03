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
data/                              data archive, one sqlite .db file per tournament
docs/                              GitHub Pages web root
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

## Workflow

1. **Archive** a tournament once it's finished. Find its host, base path and spec version, and run `utr` with that version (see [compatibility](#compatibility))

   ```
   dist/utr -version v01_09_14 -mode archive -host wbuc.wfdf.sport
   ```

   This writes `data/<slug>.db`, where `<slug>` defaults to the season id the deployment's own heartbeat reports.

   Archives are write-once: the command refuses to run if that file already exists.

2. **Publish** the archive to `docs/archive/<slug>/` for GitHub Pages, via `-mode publish` or `-mode all`. Renders every endpoint's JSON under the live API's own relative filenames, so pointing a tool's `host`/`basePath` at the published folder is a drop-in replacement for the original deployment.

3. **Update** the homepage index (planned)

   intended to be generated from what's actually in `data/*.db`, not hand-maintained, so it can never drift from the archive.

Team photos and other media event links are not currently archived at all, they are real uploaded image files served by the deployment, not JSON data, so capturing them needs a future change to fetch and store the actual assets rather than just their metadata.

## Development usage

```
just check-deps   # confirm required CLI tools are installed
just dbgen        # sqlc generate internal/db/ packages from db/ with
just build        # compile the utr binary into dist/
just clean        # remove dist/
just fmt          # go mod tidy, gofmt, golangci-lint, dprint, pg_format
```
