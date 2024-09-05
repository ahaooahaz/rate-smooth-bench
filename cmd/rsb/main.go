package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/ahaooahaz/rsb/pkg/utils/gopherlua"
	lua "github.com/yuin/gopher-lua"
)

var qps = flag.Int("qps", 10, "qps")
var d = flag.Duration("d", time.Second*5, "query duration")
var url = flag.String("url", "http://example.com", "url")
var s = flag.String("s", "", "script.lua")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *d)
	defer cancel()

	start := time.Now()

	qc := (*d).Seconds() * float64(*qps)
	sleep := float64((*d).Nanoseconds()) / qc
	var wg sync.WaitGroup

out:
	for {
		select {
		case <-ctx.Done():
			break out
		default:
			time.Sleep(time.Duration(sleep) * time.Nanosecond)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			do()
		}()

	}
	rps := float64(qc / time.Since(start).Seconds())
	wg.Wait()

	fmt.Printf("\nREQUEST COUNT: %d\nREAL QPS: %v\n", int64(qc), rps)
}

func do() {

	r, err := http.Get(*url)
	if err != nil {
		panic(err)
	}

	luaruntime := lua.NewState()
	defer luaruntime.Close()

	// 执行 Lua 文件
	if err := luaruntime.DoFile(*s); err != nil {
		panic(err)
	}

	var bodyRaw []byte
	bodyRaw, _ = io.ReadAll(r.Body)

	if luaruntime != nil {
		headers := map[string]interface{}{}
		for k := range r.Header {
			headers[k] = r.Header.Get(k)
		}

		if err := luaruntime.CallByParam(lua.P{
			Fn:      luaruntime.GetGlobal("response"),
			NRet:    1,
			Protect: true,
		}, lua.LNumber(r.StatusCode), gopherlua.GoMapToLuaTable(luaruntime, headers), lua.LString(string(bodyRaw))); err != nil {
			panic(err)
		}

	}
}
