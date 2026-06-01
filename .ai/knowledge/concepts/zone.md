# Zone

`Zone` is a `uint8` bitmask.

## Values

| Zone    | Scope            | Resources                     |
|---------|------------------|-------------------------------|
| Device  | local hardware   | memory repos, disk repos      |
| Virtual | computed/derived | archives, encryption wrappers |
| Network | remote peers     | nodes routing, gateway        |

* Device = `1` (`d`)
* Virtual = `2` (`v`)
* Network = `4` (`n`)

## Defaults

* `ZoneDefault = ZoneAll = Device|Virtual|Network`.
* `NewContext` returns a context with `ZoneDefault`.
* Apphost calls `ExcludeZone(Network)` for unauthenticated guests.
* Unauthenticated means no token or an expired token; mapped to `Anyone`.

## Context Helpers

* `WithZone(z)` replaces, `IncludeZone(z)` adds, `ExcludeZone(z)` removes,
  `LimitZone(z)` intersects.
* `ctx.Zone().Is(check)` tests that all bits in `check` are set.
