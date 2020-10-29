// Copyright 2020 songdengtao. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"net/url"
	"strings"
)

type route struct {
	pattern     string
	handlers    HandlerChain
	params      Entries
	paramKeys   []string
	paramValues []string
}

// routeValue return the value of the matched route
func routeValue(root *node, method, path string, isUnescapePathValues bool) (value *route, ok bool) {
	if value, ok = iterate(root, path, len(path), 0, ""); ok {
		if pkLen := len(value.paramKeys); pkLen > 0 {
			pvLen := len(value.paramValues)
			for i := 0; i < pkLen; i++ {
				param := Entry{Key: value.paramKeys[i]}
				if i < pvLen {
					param.Value = value.paramValues[i]
					if isUnescapePathValues {
						if paramValue, err := url.QueryUnescape(param.Value); err == nil {
							param.Value = paramValue
						}
					}
				}
				value.params = append(value.params, param)
			}
			value.paramKeys = value.paramKeys[0:0]
			value.paramValues = value.paramValues[0:0]
		}
	}
	return
}

// iterate 深度优先遍历前缀树，找出与path匹配的结点路径，记录该路径中的路由参数和处理函数
func iterate(root *node, path string, pathLen, startx int, parentPattern string) (value *route, matched bool) {
	value = new(route)
	if root == nil || (root.pattern == "" && len(root.children) == 0) {
		return nil, false
	}

	ok, stopx := false, startx
	if root.pattern == wildcardChar {
		if len(root.children) > 0 {
			for i := startx; !ok && i < pathLen; i++ {
				for _, v := range root.children {
					if v.pattern[0] == path[i] && !strings.Contains(path[startx:i], slashChar) {
						if !(pathLen-i > len(v.pattern) && len(v.children) == 0) {
							ok, stopx = true, i
						}
						break
					}
				}
			}
		}
		if !ok && len(root.handlers) > 0 {
			ok, stopx = true, pathLen
		}
	} else if root.pattern == routeParamIdentifierChar && strings.HasSuffix(parentPattern, slashChar) {
		ok = true
		stopx = pathLen
		if index := strings.Index(path[startx:], slashChar); index != -1 {
			stopx = startx + index
		}
		value.paramValues = append(value.paramValues, path[startx:stopx])
	} else {
		stopx = startx + len(root.pattern)
		if pathLen >= stopx && root.pattern == path[startx:stopx] {
			ok = true
		}
	}

	if ok {
		value.pattern += root.pattern
		value.handlers = combineHandlers(value.handlers, root.middlewares)
		if stopx == pathLen && len(root.handlers) > 0 {
			value.paramKeys = root.paramKeys
			value.handlers = combineHandlers(value.handlers, root.handlers)
			return value, true
		}

		for _, v := range root.children {
			if val, o := iterate(v, path, pathLen, stopx, root.pattern); o {
				value.pattern += val.pattern
				value.handlers = combineHandlers(value.handlers, val.handlers)
				value.paramKeys = val.paramKeys
				value.paramValues = append(value.paramValues, val.paramValues...)
				return value, true
			}
		}
	}

	return nil, false
}

// combine return the merged handlers of handlers list
func combineHandlers(list ...HandlerChain) HandlerChain {
	listLen := len(list)
	if listLen == 0 {
		return HandlerChain{}
	}

	length, lens := 0, make([]int, listLen)
	for i := 0; i < listLen; i++ {
		lens[i] = length
		length += len(list[i])
	}

	handlers := make(HandlerChain, length)
	for i := 0; i < listLen; i++ {
		copy(handlers[lens[i]:], list[i])
	}

	return handlers
}
