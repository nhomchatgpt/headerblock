# Header Block

Header Block is a middleware plugin for [Traefik](https://github.com/traefik/traefik) to block request by headers which regex matched by their name and/or value

## Configuration

### Static

```yaml
pilot:
  token: "xxxxx"

experimental:
  plugins:
    headerblock:
      moduleName: "github.com/wzator/headerblock"
      version: "v0.0.2"
```

### Docker-Compose

```yaml
      - "--experimental.plugins.headerblock.modulename=github.com/wzator/headerblock"
      - "--experimental.plugins.headerblock.version=v0.0.2"
```

### Dynamic

```yaml
http:
  middlewares:
    headerblock:
      plugin:
        headerblock:
          requestHeaders:
            - name: "name"
              value: "value"
```

### Example headerblock.yaml

```yaml
http:
  middlewares:
    headerblock:
      plugin:
        headerblock:
          requestHeaders:
            - name: "User-Agent"
              value: "MJ12bot"
            - name: "User-Agent"
              value: "Amazonbot"
            - name: "User-Agent"
              value: "SemrushBot"
            - name: "User-Agent"
              value: "Applebot"
            - name: "User-Agent"
              value: "AhrefsBot"
```

### Example docker-compose.yml

```yaml
      # Settle the ports for the entry points
      - "--entrypoints.web.address=:80"
      - "--entrypoints.web-secure.address=:443"
      - "--entrypoints.web-secure.http.middlewares=headerblock@file${TRAEFIK_PLUGINS:-}"
      - "--experimental.plugins.headerblock.modulename=github.com/wzator/headerblock"
      - "--experimental.plugins.headerblock.version=v0.0.2"
```

