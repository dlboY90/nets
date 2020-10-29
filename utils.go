// Copyright 2020 songdengtao. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

// IndexOf return the index elem of the interface{} slice
func IndexOf(data []interface{}, index int, defaultv interface{}) interface{} {
	if length := len(data); index < length {
		return data[index]
	}
	return defaultv
}

// IndexOfStrings return the index elem of the string slice
func IndexOfStrings(data []string, index int, defaultv string) string {
	if length := len(data); index < length {
		return data[index]
	}
	return defaultv
}

// IndexOfStringArrays return the index elem of the string slice list
func IndexOfStringArrays(data [][]string, index int, defaultv []string) []string {
	if length := len(data); index < length {
		return data[index]
	}
	return defaultv
}

// IndexOfStringMapArray return the index elem of the string map array list
func IndexOfStringMapArray(data []map[string]string, index int, defaultv map[string]string) map[string]string {
	if length := len(data); index < length {
		return data[index]
	}
	return defaultv
}
