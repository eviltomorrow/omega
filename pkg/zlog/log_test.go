// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zlog

import (
	"fmt"
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	global, prop, err := InitLogger(&Config{
		Level:            "info",
		Format:           "text",
		DisableTimestamp: false,
		File: FileLogConfig{
			Filename: "/tmp/log/zlog/data.log",
			MaxSize:  300,
		},
	})
	if err != nil {
		fmt.Printf("配置日志信息错误，nest error: %v\r\n", err)
		os.Exit(1)
	}
	ReplaceGlobals(global, prop)

	Info("this is shepard")
}
