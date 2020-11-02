// Copyright 2020 songdengtao. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"encoding/json"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	abortIndex int8 = math.MaxUint8 / 2
	// ContentTypeJSON json Content-Type
	ContentTypeJSON string = "application/json; charset=UTF-8"
)

// Result api response result
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Any    `json:"data"`
}

// trace context trace
type trace struct {
	Env       string    // runtime env
	StartTime time.Time // context start time
	EndTime   time.Time // context end time
	Cost      int64     // response time (microtime)
	ClientIP  string    // client ip
	Params    Any       // request params
	Status    int       // http status code
	Code      int       // result code
	Message   string    // result message
	Data      Any       // result data
	Stack     []byte    // error statck
}

// Context net context
type Context struct {
	server           *Server
	Request          *http.Request
	responser        responser
	index            int8                         // 下标，记录已执行到的位置
	handlers         HandlerChain                 // 处理函数数组
	params           Entries                      // 路由参数
	Keys             map[string]interface{}       // 请求上下文KV
	keysrw           sync.RWMutex                 // Keys读写锁
	formCacheMaps    map[string]map[string]string // 缓存表单参数
	formCacheSlices  url.Values                   // 缓存表单参数
	queryCacheMaps   map[string]map[string]string // 缓存请求参数
	queryCacheSlices url.Values                   // 缓存请求参数
	Trace            trace                        // context trace data
}

// newContext reutrn new *context
func newContext(s *Server) *Context {
	return &Context{
		server: s,
		params: make(Entries, 0),
	}
}

// init init context
func (c *Context) init(w http.ResponseWriter, req *http.Request) {
	c.Request = req
	c.responser.reset(w)
	if c.server.Config.trace {
		c.Trace = trace{
			Env:       c.server.Config.env,
			StartTime: time.Now(),
			ClientIP:  c.ClientIP(),
			Params:    c.Forms(),
		}
	}
}

// reset reset context
func (c *Context) reset() {
	c.index = -1
	c.Keys = nil
	c.handlers = nil
	c.params = nil
	c.queryCacheMaps = nil
	c.queryCacheSlices = nil
	c.formCacheMaps = nil
	c.formCacheSlices = nil
	c.Trace = trace{}
}

// Next 按index执行handler
func (c *Context) Next() {
	c.index++
	for int(c.index) < len(c.handlers) {
		c.handlers[c.index](c)
		if !c.IsAborted() {
			c.index++
		}
	}
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// Abort 终止执行后续handler
func (c *Context) Abort() {
	c.index = abortIndex
}

// AbortStatus 终止执行后续handler，并设置http status
func (c *Context) AbortStatus(code int) {
	c.Abort()
	c.responser.WriteHeader(code)
	c.responser.WriteHeaderNow()
	if c.server.Config.trace {
		c.Trace.Status = code
	}
}

// AbortJSON 终止执行后续handler，并响应输出json格式数据
func (c *Context) AbortJSON(code int, data Result) {
	c.Abort()
	c.JSON(code, data)
}

// JSON 响应输出json格式数据
func (c *Context) JSON(code int, data Result) (err error) {
	if data.Data == nil {
		// Set the default value of API Result Data to empty AnyMap
		data.Data = AnyMap{}
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	c.SetResponseHeader("Content-Type", ContentTypeJSON)
	c.responser.WriteHeader(code)
	c.responser.WriteHeaderNow()
	if _, err = c.responser.Write(bytes); err != nil {
		panic("cannot write message to writer during serve error: " + err.Error())
	}

	if c.server.Config.trace {
		c.Trace.Status = code
		c.Trace.Code = data.Code
		c.Trace.Message = data.Message
		if c.server.Config.recordResultData {
			c.Trace.Data = data.Data
		}
	}

	return
}

// Param 返回路由参数key的值，若key不存在，则第二个返回值为false
func (c *Context) Param(key string) (string, bool) {
	return c.params.Get(key)
}

// ParamMust 返回路由参数key的值
func (c *Context) ParamMust(key string) string {
	return c.params.Key(key)
}

// Params 返回全部路由参数
func (c *Context) Params() Entries {
	return c.params
}

// Query 返回请求参数key的值，若key不存在，则第二个返回值为false
func (c *Context) Query(key string) (string, bool) {
	if values, ok := c.QueryArray(key); ok {
		return values[0], true
	}
	return "", false
}

// QueryMust 返回请求参数key的值
func (c *Context) QueryMust(key string, defaults ...string) string {
	if value, ok := c.Query(key); ok {
		return value
	}
	return IndexOfStrings(defaults, 0, "")
}

// QueryArray 返回请求参数key的值(切片)，若key不存在，则第二个返回值为false
func (c *Context) QueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	if values, ok := c.queryCacheSlices[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// QueryMustArray 返回请求参数key的值(切片)
func (c *Context) QueryMustArray(key string, defaults ...[]string) []string {
	if values, ok := c.QueryArray(key); ok {
		return values
	}
	return IndexOfStringArrays(defaults, 0, nil)
}

// QueryMap 返回请求参数key的值(字典)，若key不存在，则第二个返回值为false
func (c *Context) QueryMap(key string) (map[string]string, bool) {
	if values, ok := c.queryCacheMaps[key]; ok && len(values) > 0 {
		return values, true
	}

	c.initQueryCache()
	dicts, ok := make(map[string]string), false
	for k, v := range c.queryCacheSlices {
		if i := strings.IndexByte(k, '['); i > 0 {
			if j := strings.IndexByte(k, ']'); j == len(k)-1 {
				dicts[k[i+1:j]], ok = v[0], true
			}
		}
	}

	if ok {
		if c.queryCacheMaps == nil {
			c.queryCacheMaps = make(map[string]map[string]string)
		}
		c.queryCacheMaps[key] = dicts
	}

	return dicts, ok
}

// QueryMustMap 返回请求参数key的值(字典)
func (c *Context) QueryMustMap(key string, defaults ...map[string]string) map[string]string {
	if values, ok := c.QueryMap(key); ok {
		return values
	}
	return IndexOfStringMapArray(defaults, 0, nil)
}

// Queries 返回全部查询参数
func (c *Context) Queries() url.Values {
	c.initQueryCache()
	return c.queryCacheSlices
}

// initQueryCache 初始化请求参数缓存
func (c *Context) initQueryCache() {
	if c.queryCacheSlices == nil {
		c.queryCacheSlices = c.Request.URL.Query()
	}
}

// Form 返回表单参数key的值，若key不存在，则第二个返回值为false
func (c *Context) Form(key string) (string, bool) {
	if values, ok := c.FormArray(key); ok {
		return values[0], true
	}
	return "", false
}

// FormMust 返回表单参数key的值
func (c *Context) FormMust(key string, defaults ...string) string {
	if value, ok := c.Form(key); ok {
		return value
	}
	return IndexOfStrings(defaults, 0, "")
}

// FormArray 返回表单参数key的值(切片)，若key不存在，则第二个返回值为false
func (c *Context) FormArray(key string) ([]string, bool) {
	c.initFormCache()
	if values, ok := c.formCacheSlices[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// FormMustArray 返回表单参数key的值(切片)
func (c *Context) FormMustArray(key string, defaults ...[]string) []string {
	if values, ok := c.FormArray(key); ok {
		return values
	}
	return IndexOfStringArrays(defaults, 0, nil)
}

// FormMap 返回表单参数key的值(字典)，若key不存在，则第二个返回值为false
func (c *Context) FormMap(key string) (map[string]string, bool) {
	if values, ok := c.formCacheMaps[key]; ok && len(values) > 0 {
		return values, true
	}

	c.initFormCache()
	dicts, ok := make(map[string]string), false
	for k, v := range c.formCacheSlices {
		if i := strings.IndexByte(k, '['); i > 0 {
			if j := strings.IndexByte(k, ']'); j == len(k)-1 {
				dicts[k[i+1:j]], ok = v[0], true
			}
		}
	}

	if ok {
		if c.formCacheMaps == nil {
			c.formCacheMaps = make(map[string]map[string]string)
		}
		c.formCacheMaps[key] = dicts
	}

	return dicts, ok
}

// FormMustMap 返回表单参数key的值(字典)
func (c *Context) FormMustMap(key string, defaults ...map[string]string) map[string]string {
	if values, ok := c.FormMap(key); ok {
		return values
	}
	return IndexOfStringMapArray(defaults, 0, nil)
}

// Forms 返回全部表单数据
func (c *Context) Forms() url.Values {
	c.initFormCache()
	return c.formCacheSlices
}

// initFormCache 初始化表单参数缓存
func (c *Context) initFormCache() {
	if c.formCacheSlices == nil {
		c.formCacheSlices = make(url.Values)
		if err := c.Request.ParseMultipartForm(c.server.Config.multipartMemoryMax); err != nil {
			if err != http.ErrNotMultipart {
				debugPrintf("error on parse multipart form array: %v", err)
			}
		}
		c.formCacheSlices = c.Request.PostForm
	}
}

// Set add or update Keys
func (c *Context) Set(key string, value interface{}) {
	c.keysrw.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
	c.keysrw.Unlock()
}

// Get return the value of the key in Keys
func (c *Context) Get(key string) (interface{}, bool) {
	c.keysrw.RLock()
	value, ok := c.Keys[key]
	c.keysrw.RUnlock()
	return value, ok
}

// Key return the value of the key in Keys
func (c *Context) Key(key string, defaults ...interface{}) interface{} {
	if value, ok := c.Get(key); ok {
		return value
	}
	return IndexOf(defaults, 0, nil)
}

// SetResponseHeader 设置请求头
func (c *Context) SetResponseHeader(key, value string) {
	if value == "" {
		c.responser.Header().Del(key)
		return
	}
	c.responser.Header().Set(key, value)
}

// ClientIP return the request client ip
func (c *Context) ClientIP() string {
	if c.server.Config.forwardedByClientIP {
		xForwardedFor := c.Request.Header.Get("X-Forwarded-For")
		ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
		if ip == "" {
			ip = strings.TrimSpace(c.Request.Header.Get("X-Real-Ip"))
		}
		if ip != "" {
			return ip
		}
	}

	hostport := strings.TrimSpace(c.Request.RemoteAddr)
	if ip, _, err := net.SplitHostPort(hostport); err == nil {
		return ip
	}

	return ""
}
