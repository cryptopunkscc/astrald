# apphost-js

JavaScript client for the [apphost](../../mod/apphost/README.md) WebSocket
protocol. Single ESM file, zero dependencies. Works in browsers and Node 21+.

For older Node:
```js
import { WebSocket } from 'ws'
globalThis.WebSocket = WebSocket
```

## Send a query

```js
import { connect } from './lib/apphost-js/index.js'

const host = await connect('ws://127.0.0.1:8624/.ws', { token: '...' })
console.log('connected to', host.alias, host.identity)

const stream = await host.query('user.info', { args: { name: 'alice' } })
for await (const { type, value } of stream) {
  console.log(type, value)
}
```

`host.query` opens its own WebSocket and tears it down when the stream closes.
Pass `target` to query a non-host identity. Stream iteration ends at the first
`eos` or when the WS closes.

## Register a service

```js
import { connect } from './lib/apphost-js/index.js'

const host = await connect('ws://127.0.0.1:8624/.ws', { token: '...' })

const reg = await host.register(host.guestID, async (q) => {
  console.log('incoming', q.caller, '→', q.query)

  if (q.query.startsWith('forbidden')) return q.reject(1)

  const s = await q.accept()
  s.send({ type: 'string8', value: 'hello, ' + q.caller })
  s.send({ type: 'eos' })
  s.close()
})

// run until Ctrl+C
process.on('SIGINT', () => { reg.unregister(); process.exit(0) })
```

The host pushes an `IncomingQuery` for each inbound query targeting the
registered identity. Accept opens a per-query WebSocket and returns a responder
`Stream`; reject closes the caller's query with the given numeric code. Either
must happen within 5 seconds or the caller sees route-not-found.

## Errors

```js
import { connect, errors } from './lib/apphost-js/index.js'

try {
  const host = await connect(url, { token: 'wrong' })
} catch (e) {
  if (e instanceof errors.AuthError) console.warn('bad token:', e.code)
  else if (e instanceof errors.ConnectError) console.warn('cannot reach host')
  else throw e
}

try {
  const stream = await host.query('foo.bar')
} catch (e) {
  if (e instanceof errors.RouteNotFound) console.warn('no handler')
  else if (e instanceof errors.QueryRejected) console.warn('rejected, code', e.code)
  else throw e
}
```

Exported error classes: `ConnectError`, `AuthError`, `QueryRejected{code}`,
`RouteNotFound`, `ProtocolError`.

## Object types you'll see

Each message is `{type, value}`. Common type names:

| type                                | value shape                                  |
|-------------------------------------|----------------------------------------------|
| `string8` / `string16` / `string32` | string                                       |
| `bytes32`                           | base64 string                                |
| `eos`                               | null (stream-end sentinel)                   |
| `ack`                               | null                                         |
| `apphost.access_token`              | `{Identity, Token, ExpiresAt}`               |
| `mod.apphost.host_info_msg`         | `{Identity, Alias}`                          |

For the full list see [`mod/apphost/protocol.md`](../../mod/apphost/protocol.md)
and the `_msg.go` files in `mod/apphost/`.
