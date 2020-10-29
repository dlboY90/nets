// Copyright 2020 dlboy(songdengtao). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"os"
	"strings"
)

const (
	// EnvDevelopment 开发环境
	EnvDevelopment string = "development"
	// EnvTest 测试环境
	EnvTest string = "test"
	// EnvGray 灰度环境
	EnvGray string = "gray"
	// EnvRelease 发布环境
	EnvRelease string = "release"
	// EnvProduction 生产环境
	EnvProduction string = "production"
	// defaultPriority handler默认优先级
	defaultPriority int = 2 << 10
	// defaultMultipartMemory multipart表单最大数据量
	defaultMultipartMemory int64 = 32 << 20 // 32 MB
	// nets env 环境变量名
	envVarName string = "NETS_ENV"
	// 默认的HTTP Server Addr
	defaultHTTPServerAddr string = ":8080"
	// 默认的HTTPS Server Addr
	defaultHTTPSServerAddr string = ":443"
)

// Configure nets config
type Configure struct {
	// 环境
	env string
	// debug状态
	debug bool
	// 是否开启trace
	trace bool
	// 默认handler优先级
	defaultPriority int
	// ForwardedByClientIP
	forwardedByClientIP bool
	// 表单数据最大内存值
	multipartMemoryMax int64
	// 是否记录trace result data
	recordResultData bool
	// 是否使用自定义recovery
	customRecovery bool
}

// newConfig return new config
func newConfig() *Configure {
	config := &Configure{
		trace:               false,
		defaultPriority:     defaultPriority,
		customRecovery:      false,
		forwardedByClientIP: true,
		multipartMemoryMax:  defaultMultipartMemory,
		recordResultData:    false,
	}
	config.SetEnv(os.Getenv(envVarName))
	return config
}

// SetEnv 设置环境
func (config *Configure) SetEnv(env string) {
	if env != "" {
		env = strings.ToLower(env)
	} else {
		env = EnvTest
	}

	switch env {
	case EnvDevelopment, EnvTest, EnvGray, "":
		config.env = env
		config.debug = true
	case EnvRelease, EnvProduction:
		config.env = env
		config.debug = false
	default:
		panic("unknown env: " + env)
	}
}

// SetMultipartMemoryMax set multipartMemoryMax
func (config *Configure) SetMultipartMemoryMax(max int64) {
	config.multipartMemoryMax = max
}

// SetRecordResultData 设置是否记录影响结果数据
func (config *Configure) SetRecordResultData(yesorno bool) {
	config.recordResultData = yesorno
}
