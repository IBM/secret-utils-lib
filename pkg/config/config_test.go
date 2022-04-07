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

package config

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestParseConfig(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	testcases := []struct {
		testcasename     string
		secretconfigpath string
		expectedError    error
	}{
		{
			testcasename:     "Valid config",
			secretconfigpath: "test-fixtures/valid/slclient.toml",
			expectedError:    nil,
		},
		{
			testcasename:     "Invalid config",
			secretconfigpath: "test-fixtures/invalid/slclient.toml",
			expectedError:    errors.New("Not nil"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.testcasename, func(t *testing.T) {
			pwd, err := os.Getwd()
			if err != nil {
				t.Errorf("Failed to get current working directory, test case parse config, error: %v", err)
			}

			filePath := filepath.Join(pwd, "..", "..", testcase.secretconfigpath)
			byteData, err := ioutil.ReadFile(filePath)
			if err != nil {
				t.Errorf("Failed to get current working directory, test case parse config, error: %v", err)
			}

			_, err = ParseConfig(logger, string(byteData))
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
