// Copyright 2020 songdengtao. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"sort"
)

// node 路由前缀树结点
type node struct {
	pattern     string       // 结点片段路由规则
	fullPattern string       // 结点完整路由规则
	handlers    HandlerChain // 私有处理函数
	middlewares HandlerChain // 共享处理函数
	paramKeys   []string     // 路由参数名称数组
	children    []*node      // 孩子结点
}

// methodTree HTTP method tree
type methodTree struct {
	method string
	root   *node
}

// methodTrees HTTP method trees
type methodTrees []methodTree

// get return the methodTree of methodTrees by method
// if not exist, make and return a new one
func (trees *methodTrees) get(method string) methodTree {
	for _, tree := range *trees {
		if tree.method == method {
			return tree
		}
	}
	return methodTree{method: method, root: &node{}}
}

// set add or update the methodTree of the methodTrees
func (trees *methodTrees) set(method string, tree methodTree) {
	for k, v := range *trees {
		if v.method == method {
			(*trees)[k] = tree
			return
		}
	}
	*trees = append(*trees, tree)
}

func createTrees(mMetas methodMetas) methodTrees {
	trees := make(methodTrees, 0, 9)
	middlewares := middlewareMetas(mMetas)
	for _, v := range mMetas {
		if v.method != methodMiddleware {
			tree := trees.get(v.method)
			createTree(tree.root, 0, "", v.metas, middlewares)
			trees.set(v.method, tree)
		}
	}
	return trees
}

// middlewareMetas 返回handlers按优先级先后排序后的中间件metas
func middlewareMetas(mMetas methodMetas) (middlewares []meta) {
	for _, v := range mMetas {
		if v.method == methodMiddleware {
			middlewares = v.metas
			for key, value := range middlewares {
				maps := map[int][]HandlerChain{}
				for _, v := range value.priorityHandlerses {
					maps[v.priority] = append(maps[v.priority], v.handlers)
				}

				priorities := []int{}
				for k := range maps {
					priorities = append(priorities, k)
				}
				sort.Ints(priorities)

				handlers := HandlerChain{}
				for _, priority := range priorities {
					for _, v := range maps[priority] {
						handlers = append(handlers, v...)
					}
				}

				middlewares[key].handlers = handlers
				middlewares[key].priorityHandlerses = nil
			}
			return
		}
	}
	return
}

func createTree(root *node, startx int, fullPattern string, routeMetas []meta, middlewareMetas []meta) {
	if mGroups, ok := groupMetas(startx, routeMetas, middlewareMetas); ok {
		// 按照优先级生成树，优先级越高越靠左
		for _, key := range sortGroupsKeys(mGroups) {
			mGroup := mGroups[key]
			middlewares := HandlerChain{}
			fpattern := fullPattern + mGroup.pattern
			for _, w := range middlewareMetas {
				if fpattern == w.pattern {
					middlewares = w.handlers
					break
				}
			}

			handlers, paramKeys := HandlerChain{}, []string{}
			for k, v := range mGroup.metas {
				if v.pattern == fpattern {
					handlers = mGroup.metas[k].handlers
					paramKeys = mGroup.metas[k].paramKeys
					break
				}
			}

			child := &node{
				pattern:     mGroup.pattern,
				fullPattern: fpattern,
				handlers:    handlers,
				middlewares: middlewares,
				paramKeys:   paramKeys,
				children:    make([]*node, 0),
			}

			debugPrintf("%+v\n", child)
			root.children = append(root.children, child)
			createTree(child, startx+len(mGroup.pattern), fpattern, mGroup.metas, middlewareMetas)
		}
	}
}

// sortGroupsKeys 按照优先级给metasGroups元素索引排序
func sortGroupsKeys(mGroups metasGroups) []int {
	// 调整pattern优先级，*优先级3，路由参数标识符:优先级2，其他优先级1
	ps, ws, rs := []int{}, []int{}, []int{}

	for key, mGroup := range mGroups {
		if mGroup.pattern == wildcardChar {
			ws = append(ws, key)
		} else if mGroup.pattern == routeParamIdentifierChar {
			rs = append(rs, key)
		} else {
			ps = append(ps, key)
		}
	}

	if len(rs) > 0 {
		ps = append(ps, rs...)
	}

	if len(ws) > 0 {
		ps = append(ps, ws...)
	}

	return ps
}

// groupMetas 根据前缀将metas分组
func groupMetas(startx int, metas []meta, middlewareMetas []meta) (mGroups metasGroups, ok bool) {
	metasLen := len(metas)
	if metasLen == 0 {
		return
	}

	chars := []byte{}
	for i := startx; i < metas[0].patternLen; i++ {
		isCommon, isAppend := isIndexCharCommon(i, metas, middlewareMetas)
		if isAppend {
			chars = append(chars, metas[0].pattern[i])
		}
		if !isCommon {
			break
		}
	}

	if len(chars) > 0 {
		mGroups.add(string(chars), metas...)
		ok = true
		return
	}

	groups := metasGroups{}
	for _, v := range metas {
		if startx < v.patternLen {
			groups.add(string(v.pattern[startx]), v)
		}
	}

	groupsLen := len(groups)
	if groupsLen == 0 {
		return
	}

	if groupsLen == 1 {
		mGroups, ok = groups, true
		return
	}

	for _, v := range groups {
		if data, ok := groupMetas(startx, v.metas, middlewareMetas); ok {
			for _, vv := range data {
				mGroups.add(vv.pattern, vv.metas...)
			}
		}
	}

	if len(mGroups) > 0 {
		ok = true
	}

	return
}

func isIndexCharCommon(index int, metas []meta, middlewareMetas []meta) (bool, bool) {
	for k, v := range metas {
		if metas[k].patternLen <= index {
			return false, false
		}

		char := v.pattern[index]
		if char == wildcardByte || char != metas[0].pattern[index] {
			return false, false
		}

		if index > 0 && char == routeParamIdentifierByte && v.pattern[index-1] == slashByte {
			return false, false
		}

		for _, vv := range middlewareMetas {
			if v.pattern[:index+1] == vv.pattern {
				return false, true
			}
		}
	}
	return true, true
}
