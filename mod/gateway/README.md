# gateway

## Configuration

`gateway.yaml`:

```yaml
gateway:
  enabled: true
  listen:
    - tcp:<public-ip>:6000       

visibility: public  
init_conns: 1   
max_conns: 8 
```
