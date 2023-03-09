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

package token

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCheckTokenLifeTime(t *testing.T) {
	// Creating test logger
	logger, teardown := GetTestLogger(t)
	defer teardown()

	testcases := []struct {
		testcasename  string
		token         string
		expectedError error
	}{
		{
			testcasename:  "Empty token string",
			token:         "",
			expectedError: errors.New("empty token string"),
		},
		{
			testcasename:  "Invalid token string",
			token:         "Invalid",
			expectedError: errors.New("token contains an invalid number of segments"),
		},
		{
			testcasename:  "Expired token",
			token:         "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjE2NDUwMzI5MTUsImlhdCI6MTY0NTAzMjU3NH0.P4yzEttdMsKXLNesMJPZNeoIAl93b5LTX2Xf7rJtZ4o",
			expectedError: errors.New("Token is expired"),
		},
		{
			testcasename:  "Valid token string without expiry time",
			token:         "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE2NDUwMzI1MTV9.KdjutwIasbBXwTpmNGx250t6GhiqR83Aqhxo-gPRJ5A",
			expectedError: errors.New("unable to find expiry time of token"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.testcasename, func(t *testing.T) {
			_, err := CheckTokenLifeTime(testcase.token, logger)
			if testcase.expectedError != nil {
				assert.Contains(t, err.Error(), testcase.expectedError.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// GetTestLogger ...
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
