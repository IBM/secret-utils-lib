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

package utils

const (
	// ErrCredentialsFileUndefined ...
	ErrCredentialsFileUndefined = "ibmcloud credentials file path undefined"

	// ErrCredentialsUndefined ...
	ErrCredentialsUndefined = "ibmcloud credentials undefined"

	// ErrAuthTypeUndefined ...
	ErrAuthTypeUndefined = "IBMCLOUD_AUTHTYPE undefined"

	// ErrUnknownCredentialType ...
	ErrUnknownCredentialType = "unknown IBMCLOUD_AUTHTYPE"

	// ErrAPIKeyNotProvided ...
	ErrAPIKeyNotProvided = "API key not provided"

	// ErrProfileIDNotProvided ...
	ErrProfileIDNotProvided = "Profile ID not provided"

	// APIKeyNotFound ...
	APIKeyNotFound = "api key could not be found"

	// UserNotFound ...
	UserNotFound = "user not found or active"

	// ProfileNotFound ...
	ProfileNotFound = "selected trusted profile not eligible for cr token"

	// ErrChangeInAuthType ...
	ErrChangeInAuthType = "Change in IBMCLOUD_AUTHTYPE observed"

	// ErrSecretConfigPathUndefined ...
	ErrSecretConfigPathUndefined = "SECRET_CONFIG_PATH undefined"

	// ErrEmptyAPIKey
	ErrEmptyAPIKey = "Empty API key"
)
