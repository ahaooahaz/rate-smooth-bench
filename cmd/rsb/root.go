package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"os"
	"sync"
	"time"

	ihttp "github.com/ahaooahaz/rsb/internal/http"
	"github.com/ahaooahaz/rsb/pkg/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "rsb",
	Short:   "rate smooth bench",
	Long:    "",
	Example: "rsb -s ./script.lua --url http://example.com -d 5s -qps 10",
	Run:     root,
}

var (
	v                         *bool
	d                         *time.Duration
	qps                       *int64
	url, method, body, script *string
)

func init() {
	v = rootCmd.Flags().BoolP("version", "v", false, "show version")
	d = rootCmd.Flags().DurationP("duration", "d", 5*time.Second, "process duration")
	qps = rootCmd.Flags().Int64P("qps", "q", 10, "pre second quest count")
	url = rootCmd.Flags().StringP("url", "u", "http://example.com", "request url")
	method = rootCmd.Flags().StringP("method", "m", "GET", "request method")
	body = rootCmd.Flags().StringP("body", "b", "", "request body")
	script = rootCmd.Flags().StringP("script", "s", "", "lua script path")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func root(cmd *cobra.Command, args []string) {
	if *v {
		fmt.Print(version.GetFullVersionInfo())
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *d)
	defer cancel()

	start := time.Now()

	qc := (*d).Seconds() * float64(*qps)
	sleep := float64((*d).Nanoseconds()) / qc
	var wg sync.WaitGroup

	var realqc atomic.Int64
	realqc.Store(0)

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
			r := &ihttp.Request{
				Method:    *method,
				Body:      *body,
				LuaScript: *script,
			}
			realqc.Add(1)
			e := r.Do(ctx)
			if e != nil {
				fmt.Fprint(os.Stderr, e.Error())
				os.Exit(1)
			}
		}()

	}
	rps := float64(float64(realqc.Load()) / time.Since(start).Seconds())
	wg.Wait()

	fmt.Printf("\nREQUEST COUNT: %d\nREAL QPS: %v\n", int64(realqc.Load()), rps)
}
