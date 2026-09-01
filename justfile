
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
