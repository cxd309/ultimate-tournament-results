# ultimate-tournament-results

Historic archive of ultimate tournament results, recorded using the Live! by BULA API.

Live! by BULA (API documented in the repo [live-by-bula-openapi](https://github.com/cxd309/live-by-bula-openapi)) is the public results platform used by recent WFDF, EUC and BULA events. It serves each tournament's data as JSON files for as long as the event stays publicly available, then the files disappear. This repo polls known deployments while they're still up and keeps the results permanently.

## How this works

Each tournament is archived into its own SQLite file, normalized to a schema matched exactly to whichever Live! spec version that deployment runs.

There is no attempt to force one schema to cover every API version, so each spec version gets its own schema, its own generated query code, and its own `archive`/`publish` binaries, all following the same pattern:

```
db/v01_09_14/                SQL source: schema.sql + query.sql, read by sqlc
internal/db/v01_09_14/       sqlc-generated db interface package for GO
internal/store/v01_09_14/    plain-Go-typed wrapper over internal/db/v01_09_14
internal/bula/v01_09_14/     models the Live! 1.9.14 API and maps to the store
cmd/v01_09_14/archive/       CLI: fetch one deployment, write a fresh data/<slug>.db
cmd/v01_09_14/publish/       CLI: render a data/<slug>.db as static JSON files in docs/
internal/convert/            sql to plain go type conversions
data/                        data archive, one sqlite .db file per tournament
docs/                        GitHub Pages web root and location for static JSON files
```

Binary/package names are the Live! app version with each dot-separated segment zero-padded to two digits and joined by underscores (`1.9.14` -> `v01_09_14`).

## Compatibility

| Live! by BULA app version | Binary version | Status         |
| ------------------------- | -------------- | -------------- |
| 1.9.14 - 1.9.16           | v01_09_14      | In Development |
| 1.9.17                    | v01_09_17      | Planned        |
| 2.0.0                     | -              | Not Planned    |
| 3.0.6                     | v03_00_06      | Planned        |

## Workflow

1. **Archive** a tournament once it's finished. Find its host, base path and spec version and run the archiver specific for that version (see [compatibility](#compatibility))

   ```
   dist/archive-v01_09_14 -host wbuc.wfdf.sport
   ```

   This writes `data/<slug>.db`, where `<slug>` defaults to the season id the deployment's own heartbeat reports.

   Archives are write-once: the command refuses to run if that file already exists.

2. **Publish** the archive to `docs/` for GitHub Pages (planned).

3. **Update** the homepage index (planned)

   intended to be generated from what's actually in `data/*.db`, not hand-maintained, so it can never drift from the archive.

## Development usage

```
just check-deps   # confirm required CLI tools are installed
just dbgen        # sqlc generate internal/db/ packages from db/ with
just build        # compile every version's binaries into dist/
just clean        # remove dist/
just fmt          # go mod tidy, gofmt, golangci-lint, dprint, pg_format
```
