// Copyright 2020 songdengtao. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				recovery(ctx, err)
			}
		}()
		ctx.Next()
	}
}

// recovery record stack and abort
func recovery(ctx *Context, err interface{}) {
	var brokenPipe bool
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			lowerSeErr := strings.ToLower(se.Error())
			if strings.Contains(lowerSeErr, "broken pipe") || strings.Contains(lowerSeErr, "connection reset by peer") {
				brokenPipe = true
			}
		}
	}

	errStack := stack(4, err)
	if ctx.server.Config.trace {
		ctx.Trace.Stack = errStack
	}

	if ctx.server.Config.debug {
		println(string(errStack))
	}

	if brokenPipe {
		ctx.Abort()
	} else {
		ctx.AbortStatus(http.StatusInternalServerError)
	}
}

// stack 错误栈
func stack(skip int, err interface{}) []byte {
	var lines [][]byte
	var lastFile string
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "[panic] %s :\n\n", err)
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d, %s:\n", file, line, runtime.FuncForPC(pc).Name())
		if file != lastFile {
			if data, err := ioutil.ReadFile(file); err == nil {
				lines = bytes.Split(data, []byte{'\n'})
				lastFile = file
			}
		}
		fmt.Fprintf(buf, "\t%s\n", source(lines, line))
	}
	return buf.Bytes()
}

// source 获取源码
func source(lines [][]byte, n int) []byte {
	n--
	if n < 0 || n >= len(lines) {
		return []byte("???")
	}
	return bytes.TrimSpace(lines[n])
}
