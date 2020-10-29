// Copyright 2020 dlboy(songdengtao). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import "fmt"

// debugPrintf 格式化打印
func debugPrintf(format string, values ...interface{}) {
	fmt.Printf(format, values...)
}

// debugPrintError 打印错误
func debugPrintError(err error) {

}
