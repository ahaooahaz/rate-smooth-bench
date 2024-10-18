package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"

	"time"

	ihttp "github.com/ahaooahaz/rate-smooth-bench/internal/http"
	"github.com/ahaooahaz/rate-smooth-bench/pkg/utils/interaction"
	"github.com/ahaooahaz/rate-smooth-bench/pkg/version"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
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
	ArgsInst *Args = &Args{}
)

func init() {
	ArgsInst.Version = rootCmd.Flags().BoolP("version", "v", false, "show version")
	ArgsInst.Duration = rootCmd.Flags().DurationP("duration", "d", 5*time.Second, "process duration")
	ArgsInst.QPS = rootCmd.Flags().Int64P("qps", "q", 10, "pre second quest count")
	ArgsInst.URL = rootCmd.Flags().StringP("url", "u", "", "request url")
	ArgsInst.Method = rootCmd.Flags().StringP("method", "m", "GET", "request method")
	ArgsInst.Body = rootCmd.Flags().StringP("body", "b", "", "request body")
	ArgsInst.ScriptPath = rootCmd.Flags().StringP("script", "s", "", "lua script path")
	ArgsInst.HeaderRaw = rootCmd.Flags().StringArrayP("header", "H", []string{}, "request header")
}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		interaction.Exit(err)
	}
}

type Args struct {
	Version    *bool          `json:"version"`
	Duration   *time.Duration `json:"duration" validate:"required"`
	QPS        *int64         `json:"qps" validate:"gte=1"`
	URL        *string        `json:"url" validate:"required,url"`
	Method     *string        `json:"method" validate:"required,oneof=GET POST"`
	Body       *string        `json:"body"`
	ScriptPath *string        `json:"script"`
	HeaderRaw  *[]string      `json:"header"`
	Header     http.Header    `json:"-"`
	ScriptRaw  []byte         `json:"-"`
}

func (a *Args) Validate() (err error) {
	if a == nil {
		err = fmt.Errorf("args is nil")
		return
	}

	for _, h := range *a.HeaderRaw {
		kv := strings.Split(h, ":")
		if len(kv) != 2 {
			return
		}
		a.Header.Add(kv[0], kv[1])
	}

	if *ArgsInst.ScriptPath != "" && !filepath.IsAbs(*ArgsInst.ScriptPath) {
		var temppath string
		temppath, err = os.Getwd()
		if err != nil {
			return
		}

		temppath = filepath.Join(temppath, *ArgsInst.ScriptPath)
		*ArgsInst.ScriptPath, err = filepath.Abs(temppath)
		if err != nil {
			return
		}

		var content []byte
		content, err = os.ReadFile(*ArgsInst.ScriptPath)
		if err != nil {
			return
		}

		ArgsInst.ScriptRaw = content
	}

	validate := validator.New()
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)
	validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is required!", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		return field.Tag.Get("json") // 使用json标签作为字段名
	})

	err = validate.Struct(a)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return errors.New(err.Translate(trans))
		}
	}
	return
}

func root(cmd *cobra.Command, args []string) (err error) {
	if *ArgsInst.Version {
		fmt.Print(version.GetFullVersionInfo())
		return
	}

	err = ArgsInst.Validate()
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), *ArgsInst.Duration)
	defer cancel()

	start := time.Now()

	qc := (*ArgsInst.Duration).Seconds() * float64(*ArgsInst.QPS)
	sleep := float64((*ArgsInst.Duration).Nanoseconds()) / qc
	u, err := url.Parse(*ArgsInst.URL)
	if err != nil {
		return
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
				ID:           reqid.Add(1),
				Method:       *ArgsInst.Method,
				Body:         *ArgsInst.Body,
				Header:       ArgsInst.Header,
				LuaScriptRaw: string(ArgsInst.ScriptRaw),
				URL:          u,
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

	fmt.Printf("REQUEST COUNT: %d\nREAL DURATION: %v\nREAL QPS: %v\n", int64(realqc.Load()), dur.String(), rps)
	return
}
