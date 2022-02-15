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

// Package authenticator ...
package authenticator

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/IBM/secret-utils-lib/pkg/config"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
)

// defaultSecret is the default api key or profile ID fetched from the secret
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
			logger.Error("Error parsing credentials in IBMCLOUD_CREDENTIALS_FILE", zap.Error(err))
			return nil, "", err
		}
		var authenticator Authenticator
		credentialType, _ := credentialsmap[utils.IBMCLOUD_AUTHTYPE]
		switch credentialType {
		case utils.IAM:
			defaultSecret, _ = credentialsmap[utils.IBMCLOUD_APIKEY]
			authenticator = NewIamAuthenticator(defaultSecret, logger)
		case utils.PODIDENTITY:
			defaultSecret, _ = credentialsmap[utils.IBMCLOUD_PROFILEID]
			authenticator = NewComputeIdentityAuthenticator(defaultSecret, logger)
		}
		logger.Info("Successfully initialized authenticator")
		return authenticator, credentialType, nil
	}

	logger.Error("IBMCLOUD_CREDENTIALS_FILE undefined, trying to read storage secret store")
	conf, err := config.ReadConfig(logger)
	if err != nil {
		logger.Error("Error reading secret config", zap.Error(err))
		return nil, "", err
	}

	if conf.VPC.G2APIKey == "" {
		logger.Error("Empty api key read from the secret", zap.Error(err))
		return nil, "", errors.New(utils.ErrAPIKeyNotProvided)
	}
	defaultSecret = conf.VPC.G2APIKey
	authenticator := NewIamAuthenticator(defaultSecret, logger)
	logger.Info("Successfully initialized authenticator")
	return authenticator, utils.IAM, nil
}

// parseCredentials: reads credentials and parses them into key value pairs
// a map of credentials.
func parseCredentials(logger *zap.Logger, credentialFilePath string) (map[string]string, error) {
	logger.Info("Parsing credentials", zap.String("Credential file path", credentialFilePath))

	file, err := os.Open(credentialFilePath)
	if err != nil {
		logger.Error("Unable to open the defined IBMCLOUD_CREDENTIALS_FILE", zap.String("file path", credentialFilePath))
		return nil, utils.Error{Description: "Unable to open the defined IBMCLOUD_CREDENTIALS_FILE", BackendError: err.Error()}
	}
	defer file.Close()

	// Collect the contents of the credential file in a string array.
	credentials := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		credentials = append(credentials, scanner.Text())
	}
	if len(credentials) == 0 {
		logger.Error("No credentials found", zap.String("Credentials file path", credentialFilePath))
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
		logger.Error("Credentials provided are not in the expected format", zap.String("Credentials file path", credentialFilePath))
		return nil, errors.New(utils.ErrInvalidCredentialsFormat)
	}

	// validating credentials
	credentialType, ok := credentialsmap[utils.IBMCLOUD_AUTHTYPE]
	if !ok {
		logger.Error("IBMCLOUD_AUTHTYPE is undefined", zap.String("Credentials file path", credentialFilePath))
		return nil, errors.New(utils.ErrAuthTypeUndefined)
	}

	if credentialType != utils.IAM && credentialType != utils.PODIDENTITY {
		logger.Error("Credential type provided is unknown", zap.String("Credential type", credentialType))
		return nil, errors.New(fmt.Sprintf(utils.ErrUnknownCredentialType, credentialType))
	}

	if credentialType == utils.IAM {
		if secret, ok := credentialsmap[utils.IBMCLOUD_APIKEY]; !ok || secret == "" {
			logger.Error("API key is empty")
			return nil, errors.New(utils.ErrAPIKeyNotProvided)
		}
	}

	if credentialType == utils.PODIDENTITY {
		if secret, ok := credentialsmap[utils.IBMCLOUD_PROFILEID]; !ok || secret == "" {
			logger.Error("Profile ID is empty")
			return nil, errors.New(utils.ErrProfileIDNotProvided)
		}
	}

	logger.Info("Successfully parsed credentials")
	return credentialsmap, nil
}
