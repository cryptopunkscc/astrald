# Channels

A `Channel` carries typed `Object` values over a raw stream.

* Choose the encoding (`binary`, `json`, or `text`) when constructing the channel.
* `EOS` marks the end of the stream.
* Receivers block until `EOS`, an error, or context cancellation.

## Receive Styles

| API       | Use when                                                                    |
|-----------|-----------------------------------------------------------------------------|
| `Switch`  | Client calls. Type-dispatch loop; stops on EOF, error, or helper condition. |
| `Collect` | Op handlers. Runs a callback per object; caller type-switches.              |
| `Handle`  | Subscriptions. Runs a context-cancellable loop.                             |

## Switch Helpers

| Helper                       | Effect                              |
|------------------------------|-------------------------------------|
| `channel.Expect(&ptr)`       | receive one T, stop                 |
| `channel.PassErrors`         | `*astral.ErrorMessage` → Go error   |
| `channel.BreakOnEOS`         | stop on EOS, return nil             |
| `channel.ExpectAck`          | receive Ack, stop                   |
| `channel.Collect[T](&slice)` | append all T until EOS              |
| `channel.Chan[T](ch)`        | forward T into Go channel until EOS |
| `channel.WithContext(ctx)`   | cancel Switch on ctx                |
| `channel.WithTimeout(d)`     | cancel Switch after duration        |

## Sending

* End every stream with `EOS`.
* Send mid-stream errors as `astral.Err(err)`.
* Do not signal mid-stream errors by closing the channel.

## Locked Writes

* `channel.WithLockedWrites()` wraps `Send` in a mutex so concurrent senders
  do not interleave object bytes on the underlying writer.
* Use it when multiple goroutines share one `Channel` over a single
  transport, such as the `nodes` mux multiplexing frames over one link.
* Individual `Sender` constructors do not honour this option; pass it to
  `channel.New`.
