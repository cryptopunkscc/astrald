# apphost

Client library for the [apphost](../../mod/apphost/README.md) module.

## Basic usage

```go
host := apphost.Connect(apphost.DefaultEndpoint)

fmt.Println("connected to host %v (%v)", host.HostAlias(), host.HostID())

err := host.AuthToken(token)

fmt.Println("authenticated as %v", host.GuestID())

conn, err := host.RouteQuery(query.New(nil, nil, "user.info", nil)
```