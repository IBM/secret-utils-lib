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
	"bufio"
	"errors"
	"github.com/IBM/secret-utils-lib/pkg/config"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

// defaultSecret ...
var defaultSecret string

// Authenticator ...
type Authenticator interface {
	GetToken(freshTokenRequired bool) (string, uint64, error)
	GetSecret() string
	SetSecret(secret string)
}

// NewAuthenticator initializes the particular authenticator based on the configuration provided.
func NewAuthenticator(logger *zap.Logger) (Authenticator, string, error) {
	logger.Info("Initializing authenticator")

	// Parse the file contents into name/value pairs.
	credentialFilePath := os.Getenv(utils.IBMCLOUD_CREDENTIALS_FILE)
	if credentialFilePath != "" {
		credentialsmap, err := parseCredentials(logger, credentialFilePath)
		if err != nil {
			logger.Error("Error parsing credentials", zap.Error(err))
			return nil, "", err
		}
		var authenticator Authenticator
		credentialType, _ := credentialsmap[utils.IBMCLOUD_AUTHTYPE]
		switch credentialType {
		case IAM:
			defaultSecret, _ = credentialsmap[utils.IBMCLOUD_APIKEY]
			authenticator = NewIamAuthenticator(defaultSecret, logger)
		case PODIDENTITY:
			defaultSecret, _ = credentialsmap[utils.IBMCLOUD_PROFILEID]
			authenticator = NewComputeIdentityAuthenticator(defaultSecret, logger)
		}
		logger.Info("Successfully initialized authenticator")
		return authenticator, credentialType, nil
	}

	logger.Error("IBMCLOUD_CREDENTIALS_FILE undefined")
	conf, err := config.ReadConfig(logger)
	if err != nil {
		logger.Error("Error reading secret config", zap.Error(err))
		return nil, "", err
	}

	if conf.VPCProviderConfig.G2APIKey == "" {
		logger.Error("Empty api key", zap.Error(err))
		return nil, "", utils.ErrEmptyAPIKey
	}
	defaultSecret = conf.VPCProviderConfig.G2APIKey
	authenticator = NewIamAuthenticator(defaultSecret, logger)
	return authenticator, IAM, nil
}

// parseCredentials: reads credentials and parses them into key value pairs
// a map of credentials.
func parseCredentials(logger *zap.Logger, credentialFilePath string) (map[string]string, error) {
	file, err := os.Open(credentialFilePath)
	if err != nil {
		logger.Error("Unable to open the defined IBMCLOUD_CREDENTIALS_FILE", zap.String("file path", credentialFilePath))
		return nil, err
	}
	defer file.Close()

	// Collect the contents of the credential file in a string array.
	credentials := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		credentials = append(credentials, scanner.Text())
	}
	if len(credentials) == 0 {
		return nil, errors.New(utils.ErrCredentialsUndefined)
	}

	credentialsmap := make(map[string]string)
	for _, credential := range credentials {
		if credential == "" {
			continue
		}
		// Parse the property string into name and value tokens
		var tokens = strings.SplitN(credential, "=", 2)
		if len(tokens) == 2 {
			// Store the name/value pair in the map.
			credentialsmap[tokens[0]] = tokens[1]
		}
	}

	if len(credentialsmap) == 0 {
		return nil, errors.New(utils.ErrCredentialsUndefined)
	}

	// validating credentials
	credentialType, ok := credentialsmap[IBMCLOUD_AUTHTYPE]
	if !ok {
		return nil, errors.New(utils.ErrAuthTypeUndefined)
	}

	if credentialType != utils.IAM && credentialType != utils.PODIDENTITY {
		return nil, errors.New(utils.ErrUnknownCredentialType)
	}

	if credentialType == IAM {
		if secret, ok := credentialsmap[utils.IBMCLOUD_APIKEY]; !ok || secret == "" {
			return nil, errors.New(utils.ErrAPIKeyNotProvided)
		}
	}

	if credentialType == PODIDENTITY {
		if secret, ok := credentialsmap[utils.IBMCLOUD_PROFILEID]; !ok || secret == "" {
			return nil, errors.New(utils.ErrProfileIDNotProvided)
		}
	}

	return credentialsmap, nil
}

/*
// retry ....
func retry(authenticator Authenticator, logger *zap.Logger, authType string, err error) error {
	receivedError := err
	errMsg := strings.ToLower(err.Error())
	switch authType {
	case IAM:
		if !strings.Contains(errMsg, utils.APIKeyNotFound) || !strings.Contains(errMsg, utils.UserNotFound) {
			return err
		}
	case PODIDENTITY:
		if !strings.Contains(errMsg, utils.ProfileNotFound) {
			return err
		}
	}

	for retryCount := 0; retryCount < utils.MaxRetries; retryCount++ {
		credentialsmap, err := parseCredentials(logger)
		if err != nil {
			return err
		}

		retrievedAuthType, _ := credentialsmap[IBMCLOUD_AUTHTYPE]
		if retrievedAuthType != authType {
			return errors.New(utils.ErrChangeInAuthType)
		}

		var secret string
		switch authType {
		case IAM:
			secret, _ = credentialsmap[IBMCLOUD_APIKEY]
		case PODIDENTITY:
			secret, _ = credentialsmap[IBMCLOUD_PROFILEID]
		}

		if secret == defaultSecret {
			time.Sleep(time.Second * time.Duration(utils.RetryInterval))
			continue
		}
		authenticator.SetSecret(defaultSecret)
		return nil
	}

	return receivedError
}
*/
