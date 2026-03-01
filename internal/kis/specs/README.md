# KIS Documented Snapshot

`documented_endpoints.json` is a committed snapshot of KIS documented REST specs.

- Runtime code does **not** call KIS docs site.
- Generated files are built from this snapshot only.
- CI runs `make kis-spec-check` to detect stale generated outputs.

## Refresh Flow

1. `make kis-spec-refresh`
2. Review diff (`internal/kis/specs/documented_endpoints.json`, generated `.go` files)
3. Run live smoke/contract checks for critical endpoints
4. Commit
