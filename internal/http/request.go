package http

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ahaooahaz/rate-smooth-bench/pkg/utils/gopherlua"
	"github.com/go-resty/resty/v2"
	lua "github.com/yuin/gopher-lua"
)

type Request struct {
	ID        int64
	Method    string
	URL       *url.URL
	Header    http.Header
	Body      interface{}
	LuaScript string

	_lua *lua.LState
}

func (r *Request) init() (err error) {
	if r == nil {
		return fmt.Errorf("request is nil")
	}

	if r.LuaScript != "" {
		r._lua = lua.NewState()
		err = r._lua.DoFile(r.LuaScript)
		if err != nil {
			return
		}
	}

	return
}

func (r *Request) Do(ctx context.Context) (err error) {
	if r == nil {
		return fmt.Errorf("request is nil")
	}
	err = r.init()
	if err != nil {
		return
	}
	if r._lua != nil {
		defer r._lua.Close()
	}

	header := make(map[string]string)
	for k, v := range r.Header {
		header[k] = v[0]
	}

	req := client.R().SetHeaders(header).SetBody(r.Body).SetDoNotParseResponse(true)
	var resp *resty.Response
	switch r.Method {
	case http.MethodGet:
		resp, err = req.Get(r.URL.String())
	case http.MethodPost:
		resp, err = req.Post(r.URL.String())
	default:
		return fmt.Errorf("unsupported method: %s", r.Method)
	}
	if err != nil {
		return
	}

	defer resp.RawResponse.Body.Close()
	respHeaders := map[string]interface{}{}
	for k := range resp.Header() {
		respHeaders[k] = resp.Header().Get(k)
	}

	if resp.Header().Get("Content-Type") == "text/event-stream" {
		scanner := bufio.NewScanner(resp.RawResponse.Body)
		for scanner.Scan() {
			_res := scanner.Text()
			if _res == "" {
				continue
			}
			err = r.responseSSE(r.ID, resp.StatusCode(), respHeaders, _res)
			if err != nil {
				return
			}
		}
	} else {
		var bodyRaw []byte
		bodyRaw, err = io.ReadAll(resp.RawResponse.Body)
		if err != nil {
			return
		}

		err = r.response(resp.StatusCode(), respHeaders, string(bodyRaw))
		if err != nil {
			return
		}
	}

	return
}

func (r *Request) response(statusCode int, headers map[string]interface{}, body string) (err error) {
	if r == nil || r._lua == nil {
		return
	}

	if err = r._lua.CallByParam(lua.P{
		Fn:      r._lua.GetGlobal("response"),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(statusCode), gopherlua.GoMapToLuaTable(r._lua, headers), lua.LString(body)); err != nil {
		return err
	}
	return
}

func (r *Request) responseSSE(index int64, statusCode int, headers map[string]interface{}, body string) (err error) {
	if r == nil || r._lua == nil {
		return
	}

	if err = r._lua.CallByParam(lua.P{
		Fn:      r._lua.GetGlobal("response_sse"),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(index), lua.LNumber(statusCode), gopherlua.GoMapToLuaTable(r._lua, headers), lua.LString(body)); err != nil {
		return err
	}
	return
}
