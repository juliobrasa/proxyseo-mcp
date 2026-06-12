# ProxySEO MCP Server

[Español](#español) | [English](#english)

---

## Español

Servidor MCP (Model Context Protocol) para tu servicio de proxies ProxySEO. Permite a Claude Desktop, Claude Code y cualquier cliente MCP consultar y usar tus proxies directamente desde el asistente.

### Tools disponibles

| Tool | Descripción |
|------|-------------|
| `get_proxy` | Devuelve un proxy listo para usar (URL `http://user:pass@ip:puerto` o JSON) |
| `list_proxies` | Lista todas las IPs asignadas a tu servicio (`simple` o `detailed`) |
| `proxy_status` | Estado del servicio: plan, nº de IPs, puertos, tipo de auth |
| `check_proxy` | Comprueba conectividad y anonimato a través del proxy |

Todas las tools son de **solo lectura** sobre tu propio servicio. La API key identifica tu servicio: no necesitas configurar ningún ID.

### Instalación

1. Descarga el binario para tu plataforma:
   - Linux (amd64): `proxy-mcp`
   - macOS (Apple Silicon): `proxy-mcp-darwin-arm64`
2. Dale permisos y muévelo a tu PATH:

```bash
chmod +x proxy-mcp
sudo mv proxy-mcp /usr/local/bin/proxy-mcp
```

3. Solicita tu **API key** al soporte de ProxySEO (o génerala desde el panel). Se muestra **una sola vez**: guárdala en un gestor de contraseñas.

### Configuración en Claude Desktop / Claude Code

Añade a tu `claude_desktop_config.json` (Desktop) o `.mcp.json` (Claude Code):

```json
{
  "mcpServers": {
    "proxyseo": {
      "command": "/usr/local/bin/proxy-mcp",
      "env": {
        "PROXYSEO_API_KEY": "tu-api-key-aqui"
      }
    }
  }
}
```

### Variables de entorno

| Variable | Obligatoria | Default | Descripción |
|----------|-------------|---------|-------------|
| `PROXYSEO_API_KEY` | Sí | — | API key de tu servicio (64 caracteres hex) |
| `PROXYSEO_API_URL` | No | `https://api.proxyseo.es` | URL base de la API |

### Ejemplo de uso

En Claude, simplemente pide:

> «Dame un proxy en formato URL para usar con curl»

Claude llamará a `get_proxy` y devolverá algo como:

```
http://usuario:contraseña@185.x.x.x:58542
```

> «¿Cuántas IPs tiene mi servicio y están funcionando?»

Claude combinará `proxy_status` y `check_proxy`.

### Troubleshooting

| Síntoma | Causa | Solución |
|---------|-------|----------|
| `ERROR: PROXYSEO_API_KEY environment variable is required` | Falta la variable | Añádela al bloque `env` del config MCP |
| `unauthorized: invalid or revoked API key` | Key incorrecta o revocada | Verifica la key; pide una nueva si fue revocada |
| `rate limit exceeded (60 req/min)` | Demasiadas llamadas | Espera 1 minuto; el límite es por key |
| `service not found or inactive` | Servicio suspendido/cancelado | Revisa el estado de tu servicio en el panel |
| El server no aparece en Claude | Ruta del binario incorrecta | Usa ruta absoluta en `command` y reinicia Claude |

---

## English

MCP (Model Context Protocol) server for your ProxySEO proxy service. Lets Claude Desktop, Claude Code and any MCP client query and use your proxies directly from the assistant.

### Available tools

| Tool | Description |
|------|-------------|
| `get_proxy` | Returns a ready-to-use proxy (URL `http://user:pass@ip:port` or JSON) |
| `list_proxies` | Lists all IPs assigned to your service (`simple` or `detailed`) |
| `proxy_status` | Service status: plan, IP count, ports, auth type |
| `check_proxy` | Checks connectivity and anonymity through the proxy |

All tools are **read-only** over your own service. The API key identifies your service: no service ID configuration needed.

### Installation

1. Download the binary for your platform:
   - Linux (amd64): `proxy-mcp`
   - macOS (Apple Silicon): `proxy-mcp-darwin-arm64`
2. Make it executable and move it into your PATH:

```bash
chmod +x proxy-mcp
sudo mv proxy-mcp /usr/local/bin/proxy-mcp
```

3. Request your **API key** from ProxySEO support (or generate it from the panel). It is shown **only once**: store it in a password manager.

### Claude Desktop / Claude Code configuration

Add to your `claude_desktop_config.json` (Desktop) or `.mcp.json` (Claude Code):

```json
{
  "mcpServers": {
    "proxyseo": {
      "command": "/usr/local/bin/proxy-mcp",
      "env": {
        "PROXYSEO_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

### Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PROXYSEO_API_KEY` | Yes | — | Your service API key (64 hex characters) |
| `PROXYSEO_API_URL` | No | `https://api.proxyseo.es` | API base URL |

### Usage example

In Claude, just ask:

> "Give me a proxy URL to use with curl"

Claude will call `get_proxy` and return something like:

```
http://user:password@185.x.x.x:58542
```

> "How many IPs does my service have and are they working?"

Claude will combine `proxy_status` and `check_proxy`.

### Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `ERROR: PROXYSEO_API_KEY environment variable is required` | Missing variable | Add it to the `env` block of your MCP config |
| `unauthorized: invalid or revoked API key` | Wrong or revoked key | Verify the key; request a new one if revoked |
| `rate limit exceeded (60 req/min)` | Too many calls | Wait 1 minute; the limit is per key |
| `service not found or inactive` | Service suspended/cancelled | Check your service status in the panel |
| Server not showing up in Claude | Wrong binary path | Use an absolute path in `command` and restart Claude |

### Build from source

```bash
make mcp   # produces bin/proxy-mcp (linux-amd64) and bin/proxy-mcp-darwin-arm64
```
