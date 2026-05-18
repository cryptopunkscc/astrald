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

* Default zone set: `Device|Virtual`.
* Apphost removes `Network` from anonymous guests.
* Anonymous guest means no token or an expired token, mapped to the `Anyone`
  GuestID.
