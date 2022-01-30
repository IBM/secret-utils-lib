/**
 * Copyright 2021 IBM Corp.
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

package authenticator

import (
	"github.com/IBM/secret-utils-lib/pkg/token"
	"github.com/IBM/secret-utils-lib/pkg/utils"

	"github.com/IBM/go-sdk-core/v5/core"
	"go.uber.org/zap"
)

// APIKeyAuthenticator ...
type APIKeyAuthenticator struct {
	authenticator *core.IamAuthenticator
	logger        *zap.Logger
}

// NewIamAuthenticator ...
func NewIamAuthenticator(apikey string, logger *zap.Logger) *APIKeyAuthenticator {
	logger.Info("Initializing iam authenticator")
	defer logger.Info("Initialized iam authenticator")
	aa := new(APIKeyAuthenticator)
	aa.authenticator = new(core.IamAuthenticator)
	aa.authenticator.ApiKey = apikey
	aa.logger = logger
	return aa
}

// GetToken ...
func (aa *APIKeyAuthenticator) GetToken(freshTokenRequired bool) (string, uint64, error) {
	aa.logger.Info("Fetching IAM token using api key authenticator")
	var iamtoken string
	var err error
	var tokenlifetime uint64

	if !freshTokenRequired {
		aa.logger.Info("Retreiving existing token")
		iamtoken, err = aa.authenticator.GetToken()
		if err != nil {
			aa.logger.Error("Error fetching token", zap.Error(err))
			return "", tokenlifetime, err
		}
		tokenlifetime, err = token.FetchTokenLifeTime(iamtoken)
		if err != nil {
			aa.logger.Error("Error fetching tokenlifetime", zap.Error(err))
			return "", tokenlifetime, err
		}
		if tokenlifetime > utils.TokenExpirydiff {
			aa.logger.Info("Successfully fetched IAM token")
			return iamtoken, tokenlifetime, nil
		}
	}

	tokenResponse, err := aa.authenticator.RequestToken()
	if err != nil {
		aa.logger.Error("Error fetching token", zap.Error(err))
		return "", tokenlifetime, nil
	}

	tokenlifetime, err = token.FetchTokenLifeTime(tokenResponse.AccessToken)
	if err != nil {
		aa.logger.Error("Error fetching tokenlifetime", zap.Error(err))
		return "", tokenlifetime, err
	}
	aa.logger.Info("Successfully fetched IAM token")
	return tokenResponse.AccessToken, tokenlifetime, nil
}

// GetSecret ...
func (aa *APIKeyAuthenticator) GetSecret() string {
	return aa.authenticator.ApiKey
}

// SetSecret ...
func (aa *APIKeyAuthenticator) SetSecret(secret string) {
	aa.authenticator.ApiKey = secret
}
