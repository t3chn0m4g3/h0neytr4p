<div style="text-align:center"><img src="https://github.com/t3chn0m4g3/h0neytr4p/blob/main/logo.png?raw=true" /></div>

# h0neytr4p

h0neytr4p is a lightweight HTTP/HTTPS honeypot for detecting web scanners, recon attempts, exploit probes, and known attack paths without running the real vulnerable applications behind those paths.

This fork is adjusted for [T-Pot](https://github.com/telekom-security/tpotce) style deployments and focuses on:

- Docker and Docker Compose support.
- JSON logging to a single logfile.
- Enriched request metadata such as headers, cookies, source IP, destination port, user agent details, and trap metadata.
- Multi-port trap handling.
- Payload capture for JSON, form, text, and multipart requests.
- T-Pot-compatible runtime permissions for log and payload artifacts.

Original h0neytr4p work: [pbssubhash/h0neytr4p](https://github.com/pbssubhash/h0neytr4p)

## How It Works

A trap is a JSON rule that describes:

- which request should match, for example `GET /jenkins`;
- optional headers or parameters that must match;
- which response should be returned;
- which trap name, references, and risk metadata should be written to the log.

At startup, h0neytr4p loads all `*.json` files below `traps/`. When a request matches a trap, h0neytr4p writes a JSON log entry and returns the configured response. For POST/PUT/DELETE requests, h0neytr4p can also capture request payloads and uploaded files.

Trap definitions live in:

```text
traps/
```

Response assets referenced by traps live in:

```text
traps/assets/
```

For trap authoring details, see [docs/Creating-Traps.md](docs/Creating-Traps.md).

## Requirements

For Docker deployments:

- Docker
- Docker Compose plugin

For local Go development:

- Go `1.26` or newer

The Docker build uses:

```text
golang:1.26.4-alpine3.23
```

## Quick Start With Docker

Clone and start the honeypot:

```bash
git clone https://github.com/t3chn0m4g3/h0neytr4p
cd h0neytr4p
mkdir -p log payloads
docker compose build
docker compose up
```

The default `docker-compose.yml` exposes:

| Host port | Container port | Purpose |
|---|---:|---|
| `80` | `80` | HTTP |
| `443` | `443` | HTTPS |
| `8080` | `80` | Alternative HTTP |
| `8443` | `443` | Alternative HTTPS |
| `10443` | `443` | Alternative HTTPS |

The default container command is:

```bash
./h0neytr4p \
  -cert=app.crt \
  -key=app.key \
  -log=log/log.json \
  -catchall=false \
  -payload=/opt/h0neytr4p/payloads/ \
  -wildcard=true \
  -traps=traps/
```

The image generates a self-signed certificate during build. Use `curl -k` or your browser's certificate exception flow when testing HTTPS locally.

## Runtime Paths

The compose file bind-mounts:

| Host path | Container path | Description |
|---|---|---|
| `./log/` | `/opt/h0neytr4p/log/` | JSON log output |
| `./payloads/` | `/opt/h0neytr4p/payloads/` | Captured upload payloads |

Logs are written as JSON lines to:

```text
log/log.json
```

Captured uploaded files are written to:

```text
payloads/<md5>
```

The filename is the MD5 hash of the captured file content. Runtime directories and files are chmodded to `0775` so group-based T-Pot processing can read and write them.

## Command-Line Flags

| Flag | Default in binary | Default in Docker image | Description |
|---|---|---|---|
| `-traps` | `Default` | `traps/` | Directory containing trap JSON files |
| `-log` | `Default` | `log/log.json` | JSON-lines logfile path |
| `-payload` | `Default` | `/opt/h0neytr4p/payloads/` | Directory for captured files |
| `-cert` | `Default` | `app.crt` | TLS certificate file |
| `-key` | `Default` | `app.key` | TLS key file |
| `-catchall` | `true` | `false` | Capture payloads for all requests, not only known trap paths |
| `-wildcard` | `false` | `true` | Load all traps on ports `80` and `443` |
| `-verbose` | `true` | `true` | Print log summaries to stdout |

With `-wildcard=true`, every trap is loaded on both port `80` and port `443`, regardless of the port defined in the trap file. With `-wildcard=false`, traps are loaded only on their configured `BasicInfo.Port`.

With `-catchall=false`, payloads are captured only if the request path matches at least one configured trap path. With `-catchall=true`, payloads are captured even for unmatched paths.

## Logs

Each request produces one JSON object per line. A trapped request includes fields such as:

```json
{
  "timestamp": "2026-06-08T11:31:09Z",
  "src_ip": "172.17.0.1",
  "dest_port": "8443",
  "request_method": "POST",
  "protocol": "https",
  "hostname": "127.0.0.1",
  "request_uri": "/vpns/portal/scripts/newbm.pl",
  "trapped": "true",
  "trapped_for": "CVE-2019-19781",
  "payload_hash_md5": "ba0116a9f1982d6790528a5992c0fe90",
  "payload_filename": "/opt/h0neytr4p/payloads/ba0116a9f1982d6790528a5992c0fe90"
}
```

Additional fields are added for request headers, cookies, user-agent details, trap references, and risk rating when available.

## Payload Capture

Payload capture currently handles:

- `application/json`
- `application/x-www-form-urlencoded`
- `text/plain`
- `multipart/form-data`

Size limits:

| Payload type | Limit |
|---|---:|
| Multipart | `101 KiB` |
| JSON, form, text, other body types | `11 KiB` |

If a multipart request contains a file, the file is saved under its MD5 hash in the payload directory and the log entry includes:

- `payload_hash_md5`
- `payload_filename`
- `payload_mime_type`
- `payload_parameter` for non-file multipart fields

## Creating Or Updating Traps

Minimal trap example:

```json
{
  "BasicInfo": {
    "Name": "jenkins_home",
    "Port": "443",
    "Protocol": "HTTP",
    "MitreAttackTags": "",
    "References": "",
    "RiskRating": "Critical",
    "Description": "Detect Jenkins home path probes"
  },
  "Behaviour": [
    {
      "Request": {
        "Url": "/jenkins*",
        "Method": "GET",
        "Proto": "",
        "Headers": {},
        "Params": {}
      },
      "Response": {
        "StatusCode": 302,
        "Body": "traps/assets/jenkins/default.html",
        "Headers": {},
        "Type": "file"
      },
      "trap": "true"
    }
  ]
}
```

Notes:

- `Request.Url`, `Request.Proto`, `Request.Headers`, and `Request.Params` support glob-style `*` matching.
- `Request.Proto` is optional. Leave it empty or omit it to match any HTTP version, or use values such as `HTTP/2*` for HTTP/2-specific traps.
- Use `{}` for empty headers or parameters.
- `Response.Type` can be `file` or `string`.
- For `file` responses, `Response.Body` is a path relative to the working directory.
- Restart h0neytr4p after adding or changing trap files.

More detail is available in [docs/Creating-Traps.md](docs/Creating-Traps.md).

## Tests

### Go Unit Tests

The Go tests in `pkg/` exercise the trap parser and request handler without opening real network listeners.

Run them with a local Go toolchain:

```bash
go test ./pkg
```

Or run them through the same Go Docker image used by the build:

```bash
docker run --rm -v "$PWD:/src" -w /src golang:1.26.4-alpine3.23 go test ./pkg
```

The handler tests cover:

- `text/plain` payload capture and matching against trap parameters.
- JSON payload capture and matching against top-level JSON fields.
- Multipart upload capture, MD5-based payload file naming, payload parameter logging, and T-Pot-compatible payload file permissions.

The parser test verifies that invalid trap JSON returns an error instead of terminating the process.

### Payload Smoke Test

`tests/test-cve-2019-19781-payload.sh` sends a multipart upload to the `CVE-2019-19781` trap:

```text
POST /vpns/portal/scripts/newbm.pl
```

With the default `docker-compose.yml` port mapping, start h0neytr4p and run:

```bash
tests/test-cve-2019-19781-payload.sh
```

The script verifies that:

- The trap responds with HTTP `200`.
- The uploaded file is written to the payload directory under its MD5 hash.
- The payload file mode matches the runtime mode (`0775`).
- The JSON log contains a matching `trapped=true` entry for `CVE-2019-19781`.

Useful overrides:

```bash
BASE_URL=https://127.0.0.1:8443 \
LOG_FILE=/path/to/log/log.json \
PAYLOAD_DIR=/path/to/payloads \
CONTAINER_NAME=h0neytr4p \
tests/test-cve-2019-19781-payload.sh
```

If the host-side `log/` or `payloads/` paths are not readable because of ownership or group permissions, the script falls back to `docker cp` from `CONTAINER_NAME`.

### HTTP/2 Smoke Test

`tests/test-cve-2026-23918-http2.sh` sends an HTTP/2 request to the Apache `CVE-2026-23918` trap:

```text
GET /?h0neytr4p_test=<run-id> HTTP/2
```

Run it against a TLS-enabled h0neytr4p instance:

```bash
tests/test-cve-2026-23918-http2.sh
```

The script requires a `curl` build with HTTP/2 support. It verifies that:

- The trap responds with HTTP `200`.
- The negotiated protocol is HTTP/2.
- The response body matches the Apache-style default page.
- The JSON log contains a matching `trapped=true` entry for `CVE-2026-23918` with `request_proto` starting with `HTTP/2`.

## Development Checks

Recommended checks before committing:

```bash
go test ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
find traps -name '*.json' -exec jq empty {} \;
docker compose config
docker build --progress=plain -t h0neytr4p:dev .
```

If Go is not installed locally, run Go commands through the Go Docker image:

```bash
docker run --rm -v "$PWD:/src" -w /src golang:1.26.4-alpine3.23 go test ./...
```

## Troubleshooting

### Port already in use

If `80`, `443`, or the alternate ports are already in use, change the host-side port in `docker-compose.yml`, for example:

```yaml
ports:
  - "8080:80"
  - "8443:443"
```

Then test with:

```bash
BASE_URL=https://127.0.0.1:8080 tests/test-cve-2019-19781-payload.sh
```

### Payload file not visible on the host

Check that the compose volume points to the directory you are inspecting:

```yaml
volumes:
  - ./payloads/:/opt/h0neytr4p/payloads/
```

If ownership or group permissions prevent direct host access, use the smoke test's `CONTAINER_NAME` fallback or inspect from inside the container.

### HTTPS certificate warning

The Docker image uses a self-signed certificate generated at build time. This is expected for local testing. Use `curl -k` or provide your own `-cert` and `-key` files.

## Credits

This fork is adjusted for T-Pot by [t3chn0m4g3](https://github.com/t3chn0m4g3/h0neytr4p).

Original h0neytr4p:

- Author: @pbssubhash; [Twitter](https://twitter.com/pbssubhash) | [LinkedIn](https://in.linkedin.com/in/pbssubhash)
- Rule contributor: @me-godsky; [Twitter](https://twitter.com/me_godsky) | [LinkedIn](https://in.linkedin.com/in/aakashmadaan13)
