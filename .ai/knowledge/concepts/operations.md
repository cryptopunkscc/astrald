# Operations

An Op is a named module service. A Query invokes it by method string. Ops can
be called locally, remotely, or by an App.

## Naming

* Define `Op<Name>` on `Module`.
* Expose it as `module.name`; PascalCase is converted to snake_case.
* Put the implementation in `op_<name>.go`.

## Structure

* Signature: `func(*astral.Context, *Query) error`, optionally with an args
  struct as a third parameter (`lib/routing.NewOp`).
* Mark optional fields with `query:"optional"`.
* `routing.OpRouter.AddStructPrefix(mod, "Op")` discovers `Op*` methods and
  registers them.

## Flow

* Accept the query.
* Wrap it in a Channel.
* Read or compute data.
* Stream typed Objects.
* End with `EOS` or `Ack`.

Missing required args reject before the handler runs.
