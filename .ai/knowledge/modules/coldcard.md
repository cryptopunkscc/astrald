# mod/coldcard

Provides a `crypto.Engine` that delegates BIP-137 text signing to a Coldcard hardware wallet over the `ckcc` CLI. Owns the device scan that maps attached Coldcards to public keys and the `coldcard.scan` op that re-runs that scan on demand.

## Dependencies

| Module | Why |
| --- | --- |
| `crypto` | engine auto-registered through `CryptoEngine()`; the engine returns a `MessageSigner` that signs via a USB device instead of an in-process key |
| `core/assets` | `LoadYAML` reads the (currently empty) config; `Database()` backs an unused `DB` placeholder |
| `ckcc` CLI | external binary on `$PATH`; `ckcc list`, `ckcc pubkey`, and `ckcc msg` are invoked via `os/exec` |

## Flows

- Engine registration: loader builds the `Module` -> `mod/crypto.LoadDependencies` discovers `EngineProvider`s and calls `CryptoEngine()` -> `Engine{mod}` joins the crypto engine set.
- Device discovery: `Run` launches `mod.Scan()` -> `ckcc.List()` shells out to `ckcc list` and parses `Coldcard <serial>:` lines -> for each device, `dev.PubKey(BIP44Path)` runs `ckcc -s <serial> pubkey m/1791'/0'/0'/0/0` -> store the hex pubkey in `mod.devices` keyed by serial.
- Text signer: `Engine.NewTextSigner` rejects non-`bip137` scheme and non-`secp256k1` key type -> hex-encodes the public key -> `deviceForPublicKeyHex` walks the device map to find a serial whose stored pubkey matches -> returns `MessageSigner{dev, BIP44Path}`, or `crypto.ErrUnsupported` if no device owns the key.
- Sign text: `MessageSigner.SignText` calls `dev.Msg(msg, BIP44Path)` which shells out to `ckcc -s <serial> msg -p <path> -j <msg>` -> base64-decodes the returned signature -> returns a `crypto.Signature{Scheme: "bip137", Data: sig}`.
- Op `coldcard.scan`: re-runs `mod.Scan()` and acks; used to refresh the device map after plugging in or unplugging a device.

## Source

- `mod/coldcard/module.go` - module name and `BIP44Path = "m/1791'/0'/0'/0/0"`.
- `mod/coldcard/README.md` - one-line scope statement.
- `mod/coldcard/src/loader.go`, `module.go`, `deps.go`, `config.go`, `db.go` - registration, scan-on-start lifecycle, dependency wiring, and unused config/DB placeholders.
- `mod/coldcard/src/engine.go` - `crypto.Engine` text-signer provider and the `MessageSigner` that shells out for each signature.
- `mod/coldcard/src/op_scan.go` - `coldcard.scan` handler.
- `mod/coldcard/ckcc/ckcc.go` - thin `os/exec` wrapper around the `ckcc` CLI (`List`, `Device.PubKey`, `Device.Msg`).

## Invariants

- Engine accepts only `scheme == "bip137"` and `key.Type == "secp256k1"`; everything else returns an `ErrUnsupported*` so `mod/crypto` keeps fanning out.
- Returning `crypto.ErrUnsupported` when no attached device matches the requested public key is the routing signal for "not mine, keep looking" in the fan-out.
- Signing is single-fixed-path: every signer uses `coldcard.BIP44Path`; alternative paths require code changes, not args.
- `Scan` is best-effort: device lookup failures during `dev.PubKey` are silently skipped; the device is just not added to the map.
- All device I/O is synchronous `os/exec` to the user-installed `ckcc` CLI; absence of the binary makes `List` fail and the engine register zero devices.
