# tcpfwd

This module lets you tunnel TCP connections over astral. 

## Configuration

The config file for the module is `mod_tcpfwd.yaml`.

### TCP to astral

This will listen on TCP port 8080 and tunnel incoming connections to astral
node `target` to service `http`.

```yaml
in:
  "127.0.0.1:8080": "target:http"
```

### Astral to TCP

This will register a service called `http` and forward all incoming connections
to TCP port 8080 on localhost:

```yaml
out:
  http: "127.0.0.1:8080"
```
