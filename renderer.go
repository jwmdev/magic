package main

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/ararog/timeago"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/echoview-v4"
)

//Renderer fetches the template render
func Renderer() *echoview.ViewEngine {

	gvc := goview.Config{
		Root:      "views",
		Extension: ".html",
		Master:    "layouts/master",
		Partials: []string{
			"partials/addressheader",
			"partials/walletheader",
		},
		Funcs: template.FuncMap{
			"add": func(num1, num2 int64) int64 {
				return num1 + num2
			},
			"sub": func(num1, num2 int64) int64 {
				return num1 - num2
			},

			"timeago": func(timeStamp int64) template.HTML {
				start := time.Now()
				end := time.Unix(timeStamp, 0).UTC()
				got, _ := timeago.TimeAgoWithTime(start, end)
				return template.HTML("<span class=\"tooltip is-tooltip-right\" data-tooltip=\"" + end.String() + "\">" + got + "</span>")
			},
			"timeago_raw": func(timeStamp int64) template.HTML {
				return template.HTML(time.Unix(timeStamp, 0).UTC().Format("2006-01-02 15:04"))
			},
			"timeago_max": func(timeStamp int64) template.HTML {
				start := time.Now()
				end := time.Unix(timeStamp, 0).UTC()
				got, _ := timeago.TimeAgoWithTime(start, end)
				return template.HTML(got + " (" + end.String() + ")")
			},
			"timeago_time": func(timeStamp time.Time) template.HTML {
				start := time.Now()
				got, _ := timeago.TimeAgoWithTime(start, timeStamp)
				return template.HTML("<span class=\"tooltip is-tooltip-right\" data-tooltip=\"" + timeStamp.String() + "\">" + got + "</span>")
			},
			"js": func(v interface{}) template.JS {
				a, _ := json.Marshal(v)
				return template.JS(a)
			},

			"inc": func(v int) int {
				return v + 1
			},
			"dec": func(v int) int {
				return v - 1
			},
			"inc64": func(v int64) int64 {
				return v + 1
			},
			"dec64": func(v int64) int64 {
				return v - 1
			},
		},
		DisableCache: false,
	}
	return echoview.New(gvc)

}
