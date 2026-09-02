# Spec versions with a cmd/ implementation -- extend this as new versions are added
versions := "v01_09_14 v01_09_17 v03_00_06"

# Build output directory
dist_dir := "dist"

# ── Check Dependencies: print version for all cli's required ─────────────────
check-deps:
    go version
    dprint --version
    pg_format --version
    sqlc version
    @echo "If you can read this and nothing is obviously wrong"
    @echo "All dependencies are probably found"

# ── Format: dprint, shfmt, pg_format ─────────────────────────────────────────
fmt:
    go mod tidy
    gofmt -w .
    golangci-lint run
    dprint fmt
    pg_format -i $(git ls-files '*.sql')

# ── DB Generate: sqlc generate db package  ───────────────────────────────────
dbgen:
    sqlc generate

# ── Clean: remove build output ────────────────────────────────────────────────
clean:
    rm -rf {{ dist_dir }}

# ── Build: compile cmd/<version>/<tool> into {{ dist_dir }}/<tool>-<version> ─
build: clean fmt
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p {{ dist_dir }}
    for v in {{ versions }}; do
        for dir in cmd/$v/*/; do
            tool="$(basename "$dir")"
            out="{{ dist_dir }}/${tool}-${v}"
            echo "==> building $out"
            go build -o "$out" "./$dir"
        done
    done
 