## astral

Crude go library to use astrald via its apphost module.

### Quick start

### Server
```go
port := astral.Register("myapp")

for request := range port.Next() {
    if request.Caller() == trustedID {
        conn, err := request.Accept()
        // handle the connection
        conn.Close()
    } else {
    	request.Reject()
    }
}
```

### Client
```go
conn, err := astral.Query(nodeID, "myapp")
// transfer data
conn.Close()
```