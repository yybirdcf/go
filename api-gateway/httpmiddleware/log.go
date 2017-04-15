package httpmiddleware

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/valyala/fasthttp"
)

const lowerhex = "0123456789abcdef"

type combinedLoggingHandler struct {
	writer  **os.File
	handler func(ctx *fasthttp.RequestCtx)
}

func (h *combinedLoggingHandler) ServeHTTP(ctx *fasthttp.RequestCtx) {
	t := time.Now()
	h.handler(ctx)
	writeCombinedLog(h.writer, t, fmt.Sprintf("%s", time.Since(t)), ctx)
}

func writeCombinedLog(w **os.File, ts time.Time, duration string, ctx *fasthttp.RequestCtx) {
	buf := buildCommonLogLine(ts, duration, ctx)
	buf = append(buf, ` "`...)
	buf = appendQuoted(buf, ctx.Referer())
	buf = append(buf, `" "`...)
	buf = appendQuoted(buf, ctx.UserAgent())
	buf = append(buf, `" "`...)
	buf = appendQuoted(buf, ctx.Request.Body())
	buf = append(buf, '"', '\n')
	_, err := (*w).Write(buf)
	if err != nil {
		fmt.Println(err)
	}
}

func buildCommonLogLine(ts time.Time, duration string, ctx *fasthttp.RequestCtx) []byte {
	username := "-"
	proto := "-"
	if ctx.Request.Header.IsHTTP11() {
		proto = "HTTP/1.1"
	} else {
		proto = "HTTP/1.0"
	}

	addr := ctx.RemoteIP().String()
	uri := ctx.RequestURI()
	method := ctx.Method()
	status := ctx.Response.StatusCode()
	respBodySize := "-"
	respSize := ctx.Response.Header.ContentLength()
	reqSize := ctx.Request.Header.ContentLength()
	host := ctx.Host()
	xForwardedFor := ctx.Request.Header.Peek("X-Forwarded-For")

	buf := make([]byte, 0, 3*(len(host)+len(username)+len(method)+len(uri)+len(proto)+len(host)+len(xForwardedFor)+50)/2)
	buf = append(buf, "["...)
	buf = append(buf, ts.Format("02/Jan/2006:15:04:05 -0700")...)
	buf = append(buf, `] `...)
	buf = append(buf, addr...)
	buf = append(buf, " - "...)
	buf = append(buf, username...)
	buf = append(buf, ` "`...)
	buf = append(buf, method...)
	buf = append(buf, " "...)
	buf = appendQuoted(buf, uri)
	buf = append(buf, " "...)
	buf = append(buf, proto...)
	buf = append(buf, `" `...)
	buf = append(buf, strconv.Itoa(status)...)
	buf = append(buf, " "...)
	buf = append(buf, duration...)
	buf = append(buf, " "...)
	buf = append(buf, strconv.Itoa(reqSize)...)
	buf = append(buf, " "...)
	buf = append(buf, respBodySize...)
	buf = append(buf, " "...)
	buf = append(buf, strconv.Itoa(respSize)...)
	buf = append(buf, " "...)
	buf = append(buf, host...)
	buf = append(buf, ` "`...)
	buf = append(buf, xForwardedFor...)
	buf = append(buf, `"`...)
	return buf
}

func appendQuoted(buf []byte, s []byte) []byte {
	var runeTmp [utf8.UTFMax]byte
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRune(s)
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[s[0]>>4])
			buf = append(buf, lowerhex[s[0]&0xF])
			continue
		}
		if r == rune('"') || r == '\\' { // always backslashed
			buf = append(buf, '\\')
			buf = append(buf, byte(r))
			continue
		}
		if strconv.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			buf = append(buf, runeTmp[:n]...)
			continue
		}
		switch r {
		case '\a':
			buf = append(buf, `\a`...)
		case '\b':
			buf = append(buf, `\b`...)
		case '\f':
			buf = append(buf, `\f`...)
		case '\n':
			buf = append(buf, `\n`...)
		case '\r':
			buf = append(buf, `\r`...)
		case '\t':
			buf = append(buf, `\t`...)
		case '\v':
			buf = append(buf, `\v`...)
		default:
			switch {
			case r < ' ':
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[0]>>4])
				buf = append(buf, lowerhex[s[0]&0xF])
			case r > utf8.MaxRune:
				r = 0xFFFD
				fallthrough
			case r < 0x10000:
				buf = append(buf, `\u`...)
				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			default:
				buf = append(buf, `\U`...)
				for s := 28; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			}
		}
	}
	return buf

}

func LoggingHandler(out **os.File, h func(ctx *fasthttp.RequestCtx)) func(ctx *fasthttp.RequestCtx) {
	clh := combinedLoggingHandler{
		writer:  out,
		handler: h,
	}
	return clh.ServeHTTP
}
