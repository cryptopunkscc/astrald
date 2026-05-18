# Operations

An Op is a named module service. A Query invokes it by method string. Ops can be called locally, remotely, or by an App.

## Naming

* Define `Op<Name>` on `Module`.
* Expose it as `module.name`; convert PascalCase to snake_case.
* Put the implementation in `op_<name>.go`.

## Structure

* Define an args struct for query-string parameters.
* Mark optional fields with `query:"optional"`.
* Let `ops.Set` discover the `Op*` method.

## Flow

* Accept the query.
* Wrap it in a Channel.
* Read or compute data.
* Stream typed Objects.
* End with `EOS` or `Ack`.

Missing required args reject before the handler runs.
