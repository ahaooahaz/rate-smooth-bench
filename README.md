# Rate smooth bench

Rate smooth bench is a benchmark util with smooth rate, it trigger requests smoothly within designated time and wait for all requests completely, use [lua script](/example/example.lua) to post process response, support HTTP SSE.

## Install

```bash
go install github.com/ahaooahaz/rate-smooth-bench/cmd/rsb@latest
```

## Usage

```bash
Examples:
rsb -s ./script.lua --url http://example.com -d 5s -qps 10

Flags:
  -b, --body string          request body
  -d, --duration duration    process duration (default 5s)
  -H, --header stringArray   request header
  -h, --help                 help for rsb
  -m, --method string        request method (default "GET")
  -q, --qps int              pre second quest count (default 10)
  -s, --script string        lua script path
  -u, --url string           request url (default "http://example.com")
  -v, --version              show version
```

### Post process

Use response content-type to decide post process method, when response content-type is text/event-stream, run `response_sse` function to process SSE event, otherwise run `response` function.
