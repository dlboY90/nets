// Copyright 2020 songdengtao. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

const (
	methodMiddleware string = ""
)

type priorityHandlers struct {
	handlers HandlerChain
	priority int
}

type meta struct {
	pattern            string
	patternLen         int
	priorityHandlerses []priorityHandlers
	handlers           HandlerChain
	paramKeys          []string
	index              int
}

type metasGroup struct {
	pattern string // route pattern
	metas   []meta
}

type metasGroups []metasGroup

// add create new metasGroup and add into metasGroups
func (s *metasGroups) add(pattern string, metas ...meta) {
	mGroup := s.get(pattern)
	mGroup.metas = append(mGroup.metas, metas...)
	s.set(pattern, mGroup)
}

// get return the metasGroup of metasGroups by pattern
// if not exist, make and return a new one
func (s *metasGroups) get(pattern string) metasGroup {
	for _, v := range *s {
		if v.pattern == pattern {
			return v
		}
	}
	return metasGroup{pattern: pattern}
}

// set insert or update metasGroups elem
func (s *metasGroups) set(pattern string, mGroup metasGroup) {
	for k, v := range *s {
		if v.pattern == pattern {
			(*s)[k] = mGroup
			return
		}
	}
	*s = append(*s, mGroup)
}

type methodMeta struct {
	method string
	metas  []meta
}

type methodMetas []methodMeta

// add add new route meta
// 中间件按优先级的先后执行
// 同一次注册的多个中间件优先级相同
// 对于中间件，相同的pattern，handlers会叠加
// 对于路由，相同的pattern，新handlers会替换旧的handlers
func (m *methodMetas) add(method, pattern string, priority int, paramKeys []string, handlers HandlerChain) {
	mMeta := m.get(method)

	isMiddleware := method == methodMiddleware
	meta := meta{pattern: pattern, patternLen: len(pattern), paramKeys: paramKeys}
	if isMiddleware {
		p := priorityHandlers{priority: priority, handlers: handlers}
		meta.priorityHandlerses = []priorityHandlers{p}
	} else {
		meta.handlers = handlers
	}

	if length := len(mMeta.metas); length > 0 {
		if index := indexOfMetasByPattern(mMeta.metas, pattern); index >= 0 {
			if isMiddleware {
				mMeta.metas[index].priorityHandlerses = append(mMeta.metas[index].priorityHandlerses, meta.priorityHandlerses...)
			} else {
				mMeta.metas[index].handlers = meta.handlers
			}
		} else {
			meta.index = length
			mMeta.metas = append(mMeta.metas, meta)
		}
	} else {
		meta.index = 0
		mMeta.metas = append(mMeta.metas, meta)
	}

	m.set(method, mMeta)
}

// get return the methodMeta of methodMetas by method
// if not exist, make and return a new one
func (m *methodMetas) get(method string) methodMeta {
	for _, v := range *m {
		if v.method == method {
			return v
		}
	}
	return methodMeta{method: method}
}

// set insert or update methodMetas elem
func (m *methodMetas) set(method string, mMeta methodMeta) {
	for k, v := range *m {
		if v.method == method {
			(*m)[k] = mMeta
			return
		}
	}
	*m = append(*m, mMeta)
}

// indexOfMetasByPattern return the index of meta that pattern is equal param pattern
func indexOfMetasByPattern(metas []meta, pattern string) (index int) {
	index = -1
	for _, v := range metas {
		if v.pattern == pattern {
			index = v.index
			return
		}
	}
	return
}
