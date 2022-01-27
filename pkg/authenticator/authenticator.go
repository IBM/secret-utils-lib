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
	"IBM/secret-utils-lib/pkg/utils"
	"bufio"
	"errors"
	"os"
	"strings"

	// "github.com/GunaKKIBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
)

const (
	IBMCLOUD_CREDENTIALS_FILE = "IBMCLOUD_CREDENTIALS_FILE"
	IBMCLOUD_AUTHTYPE         = "IBMCLOUD_AUTHTYPE"
	IBMCLOUD_APIKEY           = "IBMCLOUD_APIKEY"
	IBMCLOUD_PROFILEID        = "IBMCLOUD_PROFILEID"
	IAM                       = "iam"
	PODIDENTITY               = "pod-identity"
)

// Authenticator ...
type Authenticator interface {
	GetToken(freshTokenRequired bool) (string, uint64, error)
	GetSecret() string
	SetSecret(secret string)
}

// NewAuthenticator initializes the particular authenticator based on the configuration provided.
func NewAuthenticator(logger *zap.Logger) (Authenticator, string, error) {
	logger.Info("Initializing authenticator")
	credentialFilePath := os.Getenv(IBMCLOUD_CREDENTIALS_FILE)
	if credentialFilePath == "" {
		logger.Error("IBMCLOUD_CREDENTIALS_FILE undefined")
		return nil, "", errors.New(utils.ErrCredentialsFileUndefined)
	}

	file, err := os.Open(credentialFilePath)
	if err != nil {
		logger.Error("Unable to open the defined IBMCLOUD_CREDENTIALS_FILE", zap.String("file path", credentialFilePath))
		return nil, "", err
	}
	defer file.Close()

	// Collect the contents of the credential file in a string array.
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Parse the file contents into name/value pairs.
	credentialsmap, err := parseCredentials(lines)
	if err != nil {
		logger.Error("Error parsing credentials", zap.Error(err))
		return nil, "", err
	}

	var authenticator Authenticator
	credentialType, _ := credentialsmap[IBMCLOUD_AUTHTYPE]
	switch credentialType {
	case IAM:
		apiKey, _ := credentialsmap[IBMCLOUD_APIKEY]
		authenticator = NewIamAuthenticator(apiKey, logger)
	case PODIDENTITY:
		profileID, _ := credentialsmap[IBMCLOUD_PROFILEID]
		authenticator = NewComputeIdentityAuthenticator(profileID, logger)
	}
	logger.Info("Successfully initialized authenticator")
	return authenticator, credentialType, nil
}

func readCredentials()

// parseCredentials: accepts an array of strings of the form "<key>=<value>" and parses/filters them to
// a map of credentials.
func parseCredentials(credentials []string) (map[string]string, error) {
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

	if credentialType != IAM && credentialType != PODIDENTITY {
		return nil, errors.New(utils.ErrUnknownCredentialType)
	}

	if credentialType == IAM {
		if secret, ok := credentialsmap[IBMCLOUD_APIKEY]; !ok || secret == "" {
			return nil, errors.New(utils.ErrAPIKeyNotProvided)
		}
	}

	if credentialType == PODIDENTITY {
		if secret, ok := credentialsmap[IBMCLOUD_PROFILEID]; !ok || secret == "" {
			return nil, errors.New(utils.ErrProfileIDNotProvided)
		}
	}

	return credentialsmap, nil
}
