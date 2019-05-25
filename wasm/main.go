package main

import (
	"syscall/js"

	"fmt"

	"net/url"

	"strings"

	"github.com/goware/urlx"
)

var (
	markdownFormat = "[![Hits](%s)](%s)"
	showFormat     = "<a href=\"%s\"/><img src=\"%s\"/></a>"
	linkFormat     = "&lt;a href=\"%s\"/&gt;&lt;img src=\"%s\"/&gt;&lt;/a&gt;"
	incrPath       = "api/count/incr/badge.svg"
	keepPath       = "api/count/keep/badge.svg"
	defaultDomain  = ""
	defaultURL     = ""
	defaultWS      = ""
)

func parseURL(s string) (schema, host, port, path, query, fragment string, err error) {
	if s == "" {
		err = fmt.Errorf("[err] ParseURI empty uri")
	}

	url, suberr := urlx.Parse(s)
	if suberr != nil {
		err = suberr
		return
	}

	schema = url.Scheme

	host, port, err = urlx.SplitHostPort(url)
	if err != nil {
		return
	}
	if schema == "http" && port == "" {
		port = "80"
	} else if schema == "https" && port == "" {
		port = "443"
	}

	path = url.Path
	query = url.RawQuery
	fragment = url.Fragment
	return
}

func onKeyUp() {
	value := js.Global().Get("document").Call("getElementById", "badge_url").Get("value").String()
	value = strings.TrimSpace(value)
	generateBadge(value)
}

func generateBadge(value string) {
	schema, host, _, path, _, _, err := parseURL(value)
	markdown := ""
	link := ""
	show := ""
	if err != nil || (schema != "http" && schema != "https") {
		markdown = "INVALID URL"
		link = "INVALID URL"
	} else {
		normalizeURL := ""
		if path == "" || path == "/" {
			normalizeURL = fmt.Sprintf("%s://%s", schema, host)
		} else {
			normalizeURL = fmt.Sprintf("%s://%s%s", schema, host, path)
		}
		incrURL := fmt.Sprintf("%s/%s?url=%s", defaultURL, incrPath, url.QueryEscape(normalizeURL))
		keepURL := fmt.Sprintf("%s/%s?url=%s", defaultURL, keepPath, url.QueryEscape(normalizeURL))
		markdown = fmt.Sprintf(markdownFormat, incrURL, defaultURL)
		link = fmt.Sprintf(linkFormat, defaultURL, incrURL)
		show = keepURL
	}
	js.Global().Get("document").Call("getElementById", "badge_markdown").Set("innerHTML", markdown)
	js.Global().Get("document").Call("getElementById", "badge_link").Set("innerHTML", link)
	js.Global().Get("document").Call("getElementById", "badge_show").Set("src", show)
}

func registerCallbacks() {
	// It will be processing when a url input field will be received a event of keyboard up.
	js.Global().Set("generateBadge", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		onKeyUp()
		return nil
	}))

	// connect websocket
	ws := js.Global().Get("WebSocket").New(defaultWS)
	ws.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("websocket connection")
		return nil
	}))
	ws.Call("addEventListener", "close", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("websocket close")
		return nil
	}))
	ws.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		p := js.Global().Get("document").Call("createElement", "p")
		p.Set("innerHTML", args[0].Get("data"))
		js.Global().Get("document").Call("getElementById", "stream_view").Call("prepend", p)
		return nil
	}))
	ws.Call("addEventListener", "error", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("websocket error")
		return nil
	}))
}

func main() {
	println("START GO WASM")
	// LOCAL MODE
	//defaultDomain = "localhost:8080"
	//defaultURL = fmt.Sprintf("http://%s", defaultDomain)
	//defaultWS = fmt.Sprintf("ws://%s/ws", defaultDomain)

	// PRODUCTION MODE
	defaultDomain = "hits.seeyoufarm.com"
	defaultURL = fmt.Sprintf("https://%s", defaultDomain)
	defaultWS = fmt.Sprintf("wss://%s/ws", defaultDomain)
	registerCallbacks()
	c := make(chan struct{}, 0)
	<-c
}
