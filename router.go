// Copyright 2020 dlboy(songdengtao). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"net/http"
	"regexp"
)

// HandlerFunc the handler func of route
type HandlerFunc func(*Context)

// HandlerChain the set of handler func
type HandlerChain []HandlerFunc

// iRoutes routes interface
type iRoutes interface {
	Use(...HandlerFunc)                    // 中间件
	Group(string, ...HandlerFunc)          // 分组
	Handle(string, string, ...HandlerFunc) // 路由
	GET(string, ...HandlerFunc)            // GET
	POST(string, ...HandlerFunc)           // POST
	DELETE(string, ...HandlerFunc)         // DELETE
	PATCH(string, ...HandlerFunc)          // PATCH
	PUT(string, ...HandlerFunc)            // PUT
	OPTIONS(string, ...HandlerFunc)        // OPTIONs
	HEAD(string, ...HandlerFunc)           // HAED
	CONNECT(string, ...HandlerFunc)        // CONNECT
	TRACE(string, ...HandlerFunc)          // TRACE
}

// router route group manager
type router struct {
	basePath string // 基础路径<原始路径>
	server   *Server
}

// handle 注册路由和中间件
func (r *router) handle(method, relativePath string, priority int, handlers HandlerChain) {
	_, abspath, paramKeys := parseCleanPath(r.basePath, relativePath)
	r.server.metas.add(method, abspath, priority, paramKeys, handlers)
}

// handle 使用默认优先级注册路由
func (r *router) handleWithDefaultPriority(method, relativePath string, handlers HandlerChain) {
	r.handle(method, relativePath, r.server.Config.defaultPriority, handlers)
}

// Use 挂载默认优先级的中间件
func (r *router) Use(handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(methodMiddleware, r.basePath, handlers)
}

// PriorityUse 挂载自定义优先级的中间件
func (r *router) PriorityUse(priority int, handlers ...HandlerFunc) {
	r.handle(methodMiddleware, r.basePath, priority, handlers)
}

// UseRecovery 挂载recovery中间件
func (r *router) UseRecovery(handler HandlerFunc) {
	r.server.Config.customRecovery = true
	r.PriorityUse(0, handler)
}

// Group 路由组
func (r *router) Group(relativePath string, handlers ...HandlerFunc) {
	baspath, _, _ := parseCleanPath(r.basePath, relativePath)
	g := &router{basePath: baspath, server: r.server}
	g.handleWithDefaultPriority(methodMiddleware, "", handlers)
}

// Handle 注册路由
func (r *router) Handle(method, relativePath string, handlers ...HandlerFunc) {
	if matches, err := regexp.MatchString("^[A-Z]+$", method); !matches || err != nil {
		panic("http method " + method + " is not valid")
	}
	r.handleWithDefaultPriority(method, relativePath, handlers)
}

// GET GET
func (r *router) GET(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodGet, relativePath, handlers)
}

// POST POST
func (r *router) POST(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodPost, relativePath, handlers)
}

// DELETE DELETE
func (r *router) DELETE(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodDelete, relativePath, handlers)
}

// PUT PUT
func (r *router) PUT(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodPut, relativePath, handlers)
}

// PATCH PATCH
func (r *router) PATCH(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodPatch, relativePath, handlers)
}

// OPTIONS OPTIONS
func (r *router) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodOptions, relativePath, handlers)
}

// HEAD HEAD
func (r *router) HEAD(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodHead, relativePath, handlers)
}

// CONNECT CONNECT
func (r *router) CONNECT(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodConnect, relativePath, handlers)
}

// TRACE TRACE
func (r *router) TRACE(relativePath string, handlers ...HandlerFunc) {
	r.handleWithDefaultPriority(http.MethodTrace, relativePath, handlers)
}
