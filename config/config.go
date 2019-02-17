//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.
//
// some code copied from Copyright 2016 PingCAP, Inc.
// https://github.com/pingcap/tidb/blob/source-code/config/config.go
//

package config

// Config object
type Config struct {
	Host string
	Port int
}

var defaultConf = Config{
	Host: "127.0.0.1",
	Port: 4444,
}

var globalConf = defaultConf

// NewConfig create default config
func NewConfig() *Config {
	conf := defaultConf
	return &conf
}

// GetGlobalConfig returns global config
func GetGlobalConfig() *Config {
	return &globalConf
}
