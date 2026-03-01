# Kiwoom Documented Snapshot

`documented_endpoints.json` is a committed snapshot of Kiwoom documented REST specs.

- Runtime code does **not** call Kiwoom docs site.
- Generated files are built from this snapshot only.
- CI runs `make kiwoom-spec-check` to detect stale generated outputs.

## Refresh Flow

1. `make kiwoom-spec-refresh`
2. Review diff (`internal/kiwoom/specs/documented_endpoints.json`, generated `.go` files)
3. Run live smoke/contract checks for critical endpoints
4. Commit
