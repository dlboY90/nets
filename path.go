// Copyright 2020 dlboy(songdengtao). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"path"
	"strings"
)

const (
	slashChar                = "/"
	slashByte                = '/'
	wildcardChar             = "*"
	wildcardByte             = '*'
	routeParamIdentifierChar = ":"
	routeParamIdentifierByte = ':'
)

// parseCleanPath clean and parse path (path = basePath + relativePath)
// abspath 解析baspath，将路由参数名称从baspath中摘下，追加进paramKeys中并返回
// 注意：默认会去除路径中多余的斜杠
func parseCleanPath(basePath, relativePath string) (baspath, abspath string, paramKeys []string) {
	isHasSlashSuffix := strings.HasSuffix(basePath, slashChar)
	if relativePath != "" {
		isHasSlashSuffix = strings.HasSuffix(relativePath, slashChar)
	}

	path := path.Join(basePath + relativePath)
	if isHasSlashSuffix && !strings.HasSuffix(path, slashChar) {
		path += slashChar
	}

	length := len(path)
	if length == 0 {
		return
	}

	if !isHasSlashSuffix {
		path += slashChar
		length++
	}

	bytes := []byte{}
	for i := 0; i < length; i++ {
		if path[i] == slashByte || path[i] == wildcardByte {
			if i == 0 || path[i-1] != path[i] {
				bytes = append(bytes, path[i])
			}
		} else {
			bytes = append(bytes, path[i])
		}
	}
	baspath = string(bytes)

	chars := []byte{}
	for i := 0; i < length; i++ {
		if baspath[i] == routeParamIdentifierByte {
			if i > 0 && baspath[i-1] == slashByte {
				s := i
				for j := i + 1; j < length; j++ {
					if baspath[j] == slashByte {
						i = j - 1
						break
					}
				}
				if i >= s {
					paramKeys = append(paramKeys, baspath[s+1:i+1])
				}
			}
			chars = append(chars, routeParamIdentifierByte)
		} else {
			chars = append(chars, baspath[i])
		}
	}
	abspath = string(chars)

	if !isHasSlashSuffix {
		abspath = strings.TrimSuffix(abspath, slashChar)
		baspath = strings.TrimSuffix(baspath, slashChar)
	}

	return
}
