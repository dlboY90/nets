// Copyright 2020 dlboy(songdengtao). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"net/http"
	"sync"
	"time"
)

// nets functions:
// 1）RESTFul API
// 2）路由(路由参数、通配符、分组路由)
// 3）中间件（自定义全局、分组路由和单路由中间件，如全局的recovery）
// 4）支持http、https
// 5）支持请求trace

// Server Net server
type Server struct {
	router             // 路由
	Config *Configure  // 配置
	pool   sync.Pool   // the pool of nets context
	metas  methodMetas // Stores the routing registration metadata of each HTTP method
	trees  methodTrees // Stores the routing prefix tree of each HTTP method
	trace  HandlerFunc // trace handle func
}

// New return new *Server
func New() (s *Server) {
	s = &Server{Config: newConfig(), metas: make(methodMetas, 0, 10)}
	s.pool.New = func() interface{} { return newContext(s) }
	s.router = router{basePath: "/", server: s}
	return
}

// Run 将路由器连接到http.Server并开始侦听和处理HTTP请求
func (s *Server) Run(addr ...string) (err error) {
	defer func() { debugPrintError(err) }()
	s.buildTrees()
	address := IndexOfStrings(addr, 0, defaultHTTPServerAddr)
	debugPrintf("Listening and serving HTTP on %s\n", address)
	err = http.ListenAndServe(address, s)
	return
}

// RunTLS 将路由器连接到http.Server并开始侦听和处理HTTPS（安全）请求
func (s *Server) RunTLS(addr, certFile, keyFile string) (err error) {
	defer func() { debugPrintError(err) }()
	s.buildTrees()
	debugPrintf("Listening and serving HTTPS on %s\n", addr)
	err = http.ListenAndServeTLS(addr, certFile, keyFile, s)
	return
}

// Trace 注册请求结束后执行的中间件
func (s *Server) Trace(handler HandlerFunc) {
	s.Config.trace = true
	s.trace = handler
}

// ServeHTTP 实现http.Handler接口
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := s.pool.Get().(*Context)
	c.init(w, r)
	s.handleHTTPRequest(c)
	if s.trace != nil {
		s.trace(c)
	}
	c.reset()
	s.pool.Put(c)
}

// handleHTTPRequest 处理http请求
func (s *Server) handleHTTPRequest(ctx *Context) {
	defer func() {
		if s.Config.trace {
			ctx.Trace.EndTime = time.Now()
			ctx.Trace.Cost = ctx.Trace.EndTime.Sub(ctx.Trace.StartTime).Microseconds()
		}
	}()

	method, path := ctx.Request.Method, ctx.Request.URL.Path
	if route, ok := routeValue(s.trees.get(method).root, method, path, true); ok {
		ctx.params = route.params
		ctx.handlers = route.handlers
		ctx.Next()
		return
	}

	ctx.Next()
	ctx.AbortStatus(http.StatusNotFound)
}

// buildTrees 根据metas构建method前缀路由树
func (s *Server) buildTrees() {
	if !s.Config.customRecovery {
		s.PriorityUse(0, Recovery())
	}
	s.trees = createTrees(s.metas)
	s.metas = nil
}
