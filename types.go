// Copyright 2020 dlboy(songdengtao). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

// Any interface{}
type Any interface{}

// AnyMap map[string]Any
type AnyMap map[string]Any

// Entry kv string
type Entry struct {
	Key   string
	Value string
}

// Entries kv string slice
type Entries []Entry

// Key 若key在Entries中存在，则返回key的Value； 否则返回默认值
func (e Entries) Key(key string, defaults ...string) string {
	if value, ok := e.Get(key); ok {
		return value
	}
	return IndexOfStrings(defaults, 0, "")
}

// Get 若key在Entries中存在，则返回key的Value和true；否则返回字符串零值和false
func (e Entries) Get(key string) (string, bool) {
	for _, v := range e {
		if v.Key == key {
			return v.Value, true
		}
	}
	return "", false
}

// Set 若key中Entries中存在，则更新其值；否则新增key
func (e *Entries) Set(key, value string) {
	for k, v := range *e {
		if v.Key == key {
			(*e)[k].Value = value
			return
		}
	}
	*e = append(*e, Entry{Key: key, Value: value})
}
