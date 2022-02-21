/**
 * Copyright 2022 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package config ...
package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/IBM/secret-utils-lib/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestReadConfig(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()
	testcases := []struct {
		testcasename  string
		configdir     string
		expectedError error
	}{
		{
			testcasename:  "Empty secret config path",
			configdir:     "",
			expectedError: utils.Error{Description: utils.ErrSecretConfigPathUndefined},
		},
		{
			testcasename:  "Invalid secret config",
			configdir:     "test-fixtures/invalid",
			expectedError: utils.Error{Description: "Failed to read config file"},
		},
		{
			testcasename:  "Valid secret config",
			configdir:     "test-fixtures/valid",
			expectedError: nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.testcasename, func(t *testing.T) {
			pwd, err := os.Getwd()
			if err != nil {
				t.Errorf("Failed to get current working directory, test case read config, error: %v", err)
			}
			if testcase.configdir == "" {
				err = os.Setenv("SECRET_CONFIG_PATH", "")
			} else {
				err = os.Setenv("SECRET_CONFIG_PATH", filepath.Join(pwd, "..", "..", testcase.configdir))
			}
			if err != nil {
				t.Errorf("Failed to set env variable, test case related to read config will fail error: %v", err)
			}
			defer os.Unsetenv("SECRET_CONFIG_PATH")
			_, err = ReadConfig(logger)
			if testcase.expectedError != nil {
				assert.NotNil(t, err)
			}
		})
	}
}

func GetTestLogger(t *testing.T) (logger *zap.Logger, teardown func()) {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(zap.DebugLevel)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	buf := &bytes.Buffer{}

	logger = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(buf),
			atom,
		),
		zap.AddCaller(),
	)

	teardown = func() {
		_ = logger.Sync()
		if t.Failed() {
			t.Log(buf)
		}
	}
	return
}
