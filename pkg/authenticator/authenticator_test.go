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

// Package authenticator ...
package authenticator

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/IBM/secret-utils-lib/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewAuthenticator(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()
	testcases := []struct {
		testcasename            string
		ibmcloudCredentialsPath string
		secretconfigpath        string
		expectedError           error
	}{
		{
			testcasename:            "Empty config paths",
			ibmcloudCredentialsPath: "",
			secretconfigpath:        "",
			expectedError:           utils.Error{Description: utils.ErrSecretConfigPathUndefined},
		},
		{
			testcasename:            "Valid cloud credential path with api key",
			ibmcloudCredentialsPath: "test-fixtures/ibmcloud_credentials/valid/apikey.toml",
			secretconfigpath:        "",
			expectedError:           nil,
		},
		{
			testcasename:            "Valid cloud credential path with trusted profile",
			ibmcloudCredentialsPath: "test-fixtures/ibmcloud_credentials/valid/trusted_profile.toml",
			secretconfigpath:        "",
			expectedError:           nil,
		},
		{
			testcasename:            "Invalid cloud credential path",
			ibmcloudCredentialsPath: "test-fixtures/ibmcloud_credentials/non-exist.toml",
			secretconfigpath:        "",
			expectedError:           errors.New("Not nil"),
		},
		{
			testcasename:            "Invalid secret config",
			ibmcloudCredentialsPath: "",
			secretconfigpath:        "test-fixtures/invalid",
			expectedError:           errors.New("Not nil"),
		},
		{
			testcasename:            "Valid secret config",
			ibmcloudCredentialsPath: "",
			secretconfigpath:        "test-fixtures/valid",
			expectedError:           nil,
		},
		{
			testcasename:            "Empty cloud credential config",
			ibmcloudCredentialsPath: "test-fixtures/ibmcloud_credentials/invalid/empty.toml",
			secretconfigpath:        "",
			expectedError:           errors.New("Not nil"),
		},
		{
			testcasename:            "Invalid credentials entries",
			ibmcloudCredentialsPath: "test-fixtures/ibmcloud_credentials/invalid/invalidMap.toml",
			secretconfigpath:        "",
			expectedError:           errors.New("Not nil"),
		},
		{
			testcasename:            "Invalid auth type",
			ibmcloudCredentialsPath: "test-fixtures/ibmcloud_credentials/invalid/invalidauthtype.toml",
			secretconfigpath:        "",
			expectedError:           errors.New("Not nil"),
		},
		{
			testcasename:            "Api key empty",
			ibmcloudCredentialsPath: "test-fixtures/ibmcloud_credentials/invalid/emptyapikey.toml",
			secretconfigpath:        "",
			expectedError:           errors.New("Not nil"),
		},
		{
			testcasename:            "Trusted profile empty",
			ibmcloudCredentialsPath: "test-fixtures/ibmcloud_credentials/invalid/emptytrustedprofile.toml",
			secretconfigpath:        "",
			expectedError:           errors.New("Not nil"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.testcasename, func(t *testing.T) {
			pwd, _ := os.Getwd()
			if testcase.ibmcloudCredentialsPath != "" {
				_ = os.Setenv("IBMCLOUD_CREDENTIALS_FILE", filepath.Join(pwd, "..", "..", testcase.ibmcloudCredentialsPath))
				defer os.Unsetenv("IBMCLOUD_CREDENTIALS_FILE")
			}
			if testcase.secretconfigpath != "" {
				_ = os.Setenv("SECRET_CONFIG_PATH", filepath.Join(pwd, "..", "..", testcase.secretconfigpath))
				defer os.Unsetenv("SECRET_CONFIG_PATH")
			}
			_, _, err = NewAuthenticator(logger)
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
