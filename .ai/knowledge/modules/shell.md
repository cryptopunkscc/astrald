# mod/shell

Exposes a node-local interactive shell that turns terminal lines into routed queries against loaded module routers. Owns shell session parsing, scoped op dispatch for local modules, sudo identity switching for `shell.shell`, and `shell.spec` discovery output.

## Dependencies

| Module | Why |
|---|---|
| `auth` | `OpShell` authorizes `as=<alias>` with `SudoAction` before rebinding the session identity |
| `dir` | `OpShell` resolves the `as` argument to an identity |
| `objects` | injected into `Deps`; currently retained without direct calls |
| `core` | `LoadDependencies` inspects `core.Node.Modules().Loaded()` and adds every `HasRouter` module to the scope router |
| `scheduler` | `LogAction` implements `scheduler.Task` for shell-triggered logging work |

## Flows

- Scope setup: `Load` builds a shell op router -> lifts `spec` into the root router as `.spec` -> creates `ScopeRouter` -> adds the shell scope.
- Module scope discovery: `LoadDependencies` injects dependencies -> walks loaded modules -> any module implementing `Router() astral.Router` is added under its stringified module name.
- Shell session start: `shell.shell` -> resolve optional `As` -> require `auth.Authorize(SudoAction)` when present -> otherwise use `q.Caller()` -> `AcceptRaw` -> `NewSession(...).Run`.
- REPL command: prompt with guest and host IDs -> `Terminal.ReadLine` -> `shell.Split` -> empty input continues, quote mismatch prints an error, `exit` returns.
- Query launch: first word becomes op name -> remaining words become `query.ArgsToMap` -> `query.New(ctx.Identity(), node.Identity(), op, params)` -> mark `Extra["interface"]="terminal"` -> `query.RouteInFlight`.
- Query streaming: copy terminal input to the op connection through a cancelable `ContextReader` -> copy op output back to the terminal -> close op connection -> cancel stdin copy -> print `ok`.
- Spec stream: `shell.spec` opens a channel with requested output format -> sorts `mod.scopes.Spec()` by name -> applies optional exact `Op` filter -> sends each `routing.OpSpec` -> sends `EOS`.
- Local routing: `RouteQuery` accepts only queries targeted at the local node identity -> delegates to `ScopeRouter`; other targets fall through with `RouteNotFound`.

## Source

- `mod/shell/module.go`, `terminal.go` - public module constants and the line-oriented terminal wrapper.
- `mod/shell/src/loader.go`, `module.go`, `deps.go`, `config.go` - module construction, empty config, dependency injection, root `.spec` lift, and loaded-router discovery.
- `mod/shell/src/query_router.go` - local target check and scoped query dispatch.
- `mod/shell/src/op_shell.go`, `op_spec.go` - interactive shell and op-spec query handlers.
- `mod/shell/src/session.go`, `prompt.go` - REPL loop, command parsing, routed query launch, and prompt rendering.
- `mod/shell/src/log_action.go` - scheduler task used for shell log actions.

## Invariants

- `shell.shell` only changes the effective session identity when `auth.Authorize(SudoAction)` grants the requested `as` identity.
- Without `as`, the shell session runs as `q.Caller()`.
- `exit` ends the session; empty lines and quote-mismatch lines do not end it.
- `.spec` is reachable through the root scope and `shell.spec` is reachable through the shell scope.
- `shell.spec` always terminates its stream with `EOS`.
