package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ahaooahaz/rsb/pkg/utils/gopherlua"
	"github.com/go-resty/resty/v2"
	lua "github.com/yuin/gopher-lua"
)

type Request struct {
	Method    string
	URL       *url.URL
	Headers   map[string]string
	Body      interface{}
	LuaScript string
}

func (r *Request) Do(ctx context.Context) (err error) {
	if r == nil {
		return fmt.Errorf("request is nil")
	}

	req := client.R().SetHeaders(r.Headers).SetBody(r.Body)
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

	if r.LuaScript != "" {
		luaRuntime := lua.NewState()
		defer luaRuntime.Close()

		if err := luaRuntime.DoFile(r.LuaScript); err != nil {
			panic(err)
		}

		respHeaders := map[string]interface{}{}
		for k, v := range resp.Header() {
			respHeaders[k] = v
		}
		if err = luaRuntime.CallByParam(lua.P{
			Fn:      luaRuntime.GetGlobal("response"),
			NRet:    0,
			Protect: true,
		}, lua.LNumber(resp.StatusCode()), gopherlua.GoMapToLuaTable(luaRuntime, respHeaders), lua.LString(string(resp.Body()))); err != nil {
			return err
		}
	}
	return
}
