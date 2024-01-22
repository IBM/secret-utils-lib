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

	k8s_utils "github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
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

func TestFrameTokenExchangeURL(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	testcases := []struct {
		testCaseName             string
		secretDataPath           string
		clusterInfoPath          string
		expectedTokenURL         string
		providedTokenExchangeURL bool
		providerToBeUsed         string
	}{
		{
			testCaseName:             "VPC gen2 prod cluster",
			secretDataPath:           "test-fixtures/valid/vpc-gen2/prod/slclient.toml",
			clusterInfoPath:          "test-fixtures/valid/vpc-gen2/prod/cluster-info.json",
			expectedTokenURL:         utils.ProdPrivateIAMURL + tokenExchangePath,
			providedTokenExchangeURL: false,
			providerToBeUsed:         utils.Bluemix,
		},
		{
			testCaseName:             "VPC gen2 stage cluster",
			secretDataPath:           "test-fixtures/valid/vpc-gen2/stage/slclient.toml",
			clusterInfoPath:          "test-fixtures/valid/vpc-gen2/stage/cluster-info.json",
			expectedTokenURL:         utils.StagePrivateIAMURL + tokenExchangePath,
			providedTokenExchangeURL: false,
			providerToBeUsed:         utils.Bluemix,
		},
		{
			testCaseName:             "VPC gen2 dev cluster",
			secretDataPath:           "test-fixtures/valid/vpc-gen2/dev/slclient.toml",
			clusterInfoPath:          "test-fixtures/valid/vpc-gen2/dev/cluster-info.json",
			expectedTokenURL:         utils.StagePrivateIAMURL + tokenExchangePath,
			providedTokenExchangeURL: false,
			providerToBeUsed:         utils.VPC,
		},
		{
			testCaseName:             "VPC gen2 prod private endpoint provided in slclient.toml",
			secretDataPath:           "test-fixtures/valid/vpc-gen2/prod/slclient-private.toml",
			clusterInfoPath:          "test-fixtures/valid/vpc-gen2/prod/cluster-info.json",
			expectedTokenURL:         "https://private.iam.cloud.ibm.com" + tokenExchangePath,
			providedTokenExchangeURL: true,
			providerToBeUsed:         utils.VPC,
		},
		{
			testCaseName:             "Classic cluster prod",
			secretDataPath:           "test-fixtures/valid/classic/prod/slclient.toml",
			clusterInfoPath:          "test-fixtures/valid/classic/prod/cluster-info.json",
			expectedTokenURL:         "https://iam.cloud.ibm.com" + tokenExchangePath,
			providedTokenExchangeURL: true,
			providerToBeUsed:         utils.Bluemix,
		},
		{
			testCaseName:             "Classic cluster stage",
			secretDataPath:           "test-fixtures/valid/classic/stage/slclient.toml",
			clusterInfoPath:          "test-fixtures/valid/classic/stage/cluster-info.json",
			expectedTokenURL:         "https://iam.test.cloud.ibm.com" + tokenExchangePath,
			providedTokenExchangeURL: true,
			providerToBeUsed:         utils.Bluemix,
		},
		{
			testCaseName:             "Satellite cluster prod",
			secretDataPath:           "test-fixtures/valid/vpc-gen2/prod/satellite/slclient.toml",
			clusterInfoPath:          "test-fixtures/valid/vpc-gen2/prod/satellite/cluster-info.json",
			expectedTokenURL:         "https://iam.cloud.ibm.com" + tokenExchangePath,
			providedTokenExchangeURL: true,
			providerToBeUsed:         utils.VPC,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.testCaseName, func(t *testing.T) {
			pwd, err := os.Getwd()
			if err != nil {
				t.Errorf("Failed to get current working directory, test case parse config, error: %v", err)
			}

			secretfilePath := filepath.Join(pwd, "..", "..", testcase.secretDataPath)
			clusterInfoPath := filepath.Join(pwd, "..", "..", testcase.clusterInfoPath)
			k8sClient, _ := k8s_utils.FakeGetk8sClientSet()
			err = k8s_utils.FakeCreateSecret(k8sClient, utils.DEFAULT, secretfilePath)
			if err != nil {
				t.Errorf("Failed to create secret, error: %v", err)
			}
			err = k8s_utils.FakeCreateCM(k8sClient, clusterInfoPath)
			if err != nil {
				t.Errorf("Failed to create cluster info config map, error: %v", err)
			}

			returnedURL, provided := FrameTokenExchangeURL(k8sClient, testcase.providerToBeUsed, logger)
			assert.Equal(t, returnedURL, testcase.expectedTokenURL)
			assert.Equal(t, provided, testcase.providedTokenExchangeURL)
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
