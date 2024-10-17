package main

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"strings"
	"sync/atomic"

	"os"
	"time"

	ihttp "github.com/ahaooahaz/rate-smooth-bench/internal/http"
	"github.com/ahaooahaz/rate-smooth-bench/pkg/utils/interaction"
	"github.com/ahaooahaz/rate-smooth-bench/pkg/version"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var rootCmd = &cobra.Command{
	Use:     "rsb",
	Short:   "rate smooth bench",
	Long:    "",
	Example: "rsb -s ./script.lua --url http://example.com -d 5s -qps 10",
	RunE:    root,
}

var (
	v                          *bool
	d                          *time.Duration
	qps                        *int64
	urlx, method, body, script *string
	headersx                   *[]string
)

func init() {
	v = rootCmd.Flags().BoolP("version", "v", false, "show version")
	d = rootCmd.Flags().DurationP("duration", "d", 5*time.Second, "process duration")
	qps = rootCmd.Flags().Int64P("qps", "q", 10, "pre second quest count")
	urlx = rootCmd.Flags().StringP("url", "u", "http://example.com", "request url")
	method = rootCmd.Flags().StringP("method", "m", "GET", "request method")
	body = rootCmd.Flags().StringP("body", "b", "", "request body")
	script = rootCmd.Flags().StringP("script", "s", "", "lua script path")
	headersx = rootCmd.Flags().StringArrayP("header", "H", []string{}, "request header")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		interaction.Exit(err)
	}
}

func root(cmd *cobra.Command, args []string) (err error) {
	if *v {
		fmt.Print(version.GetFullVersionInfo())
		return
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), *d)
	defer cancel()

	start := time.Now()

	qc := (*d).Seconds() * float64(*qps)
	sleep := float64((*d).Nanoseconds()) / qc
	u, err := url.Parse(*urlx)
	if err != nil {
		return
	}
	headers := map[string]string{}
	for _, h := range *headersx {
		kv := strings.Split(h, ":")
		if len(kv) != 2 {
			return
		}
		headers[kv[0]] = kv[1]
	}

	if *script != "" && !filepath.IsAbs(*script) {
		var temppath string
		temppath, err = os.Getwd()
		if err != nil {
			return
		}

		temppath = filepath.Join(temppath, *script)
		*script, err = filepath.Abs(temppath)
		if err != nil {
			return
		}
	}

	var realqc, reqid atomic.Int64
	var eg errgroup.Group
	eg.SetLimit(math.MaxInt32)

out:
	for {
		select {
		case <-ctx.Done():
			break out
		default:
			time.Sleep(time.Duration(sleep) * time.Nanosecond)
		}

		eg.Go(func() error {
			r := &ihttp.Request{
				ID:        reqid.Add(1),
				Method:    *method,
				Body:      *body,
				Headers:   headers,
				LuaScript: *script,
				URL:       u,
			}
			realqc.Add(1)
			return r.Do(ctx)
		})
	}
	dur := time.Since(start)
	rps := float64(float64(realqc.Load()) / dur.Seconds())
	err = eg.Wait()
	if err != nil {
		return
	}

	fmt.Printf("REQUEST COUNT: %d\nREAL DUR: %v\nREAL QPS: %v\n", int64(realqc.Load()), dur.String(), rps)
	return
}
