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
	"github.com/IBM/secret-utils-lib/pkg/config"
	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
)

var (
	FakeAuthType string
)

// FakeAuthenticator ...
type FakeAuthenticator struct {
	secret string
	url    string
}

// FakeNewAuthenticator ...
func FakeNewAuthenticator(logger *zap.Logger) (Authenticator, string, error) {
	logger.Info("Initializing fake authenticator")
	k8s_utils.FakeAuthType = FakeAuthType

	secretData, secretname, err := k8s_utils.FakeGetSecretData(logger)
	if err != nil {
		logger.Error("Error fetching secret", zap.Error(err))
		return nil, "", err
	}

	if secretname == utils.IBMCLOUD_CREDENTIALS_SECRET {
		credentialsmap, err := parseIBMCloudCredentials(logger, secretData)
		if err != nil {
			logger.Error("Error parsing credentials", zap.Error(err))
			return nil, "", err
		}
		var authenticator Authenticator
		credentialType := credentialsmap[utils.IBMCLOUD_AUTHTYPE]
		switch credentialType {
		case utils.IAM:
			defaultSecret = credentialsmap[utils.IBMCLOUD_APIKEY]
			authenticator = NewIamAuthenticator(defaultSecret, logger)
		case utils.PODIDENTITY:
			defaultSecret = credentialsmap[utils.IBMCLOUD_PROFILEID]
			authenticator = NewComputeIdentityAuthenticator(defaultSecret, logger)
		}
		logger.Info("Successfully initialized authenticator")
		return authenticator, credentialType, nil
	}

	// Parse it the secret is storage-secret-store
	conf, err := config.ParseConfig(logger, secretData)
	if err != nil {
		logger.Error("Error parsing config", zap.Error(err))
		return nil, "", err
	}

	if conf.VPC.G2APIKey == "" {
		logger.Error("Empty api key read from the secret", zap.Error(err))
		return nil, "", utils.Error{Description: utils.ErrAPIKeyNotProvided}
	}

	defaultSecret = conf.VPC.G2APIKey
	authenticator := NewIamAuthenticator(defaultSecret, logger)
	logger.Info("Successfully initialized authenticator")
	return authenticator, utils.DEFAULT, nil
}

// GetSecret ...
func (fa *FakeAuthenticator) GetSecret() string {
	return fa.secret
}

// SetSecret ...
func (fa *FakeAuthenticator) SetSecret(secret string) {
	fa.secret = secret
}

// SetURL ...
func (fa *FakeAuthenticator) SetURL(url string) {
	fa.url = url
}

// GetToken ...
func (fa *FakeAuthenticator) GetToken(freshTokenRequired bool) (string, uint64, error) {
	if freshTokenRequired {
		return "token", 3600, nil
	}
	return "token", 1000, nil
}
