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
	// MaxRetries ...
	MaxRetries int = 4
	// RetryInterval in seconds
	RetryInterval int = 15
	// IBMCLOUD_CREDENTIALS_FILE ...
	IBMCLOUD_CREDENTIALS_FILE = "IBMCLOUD_CREDENTIALS_FILE"
	// IBMCLOUD_AUTHTYPE ...
	IBMCLOUD_AUTHTYPE = "IBMCLOUD_AUTHTYPE"
	// IBMCLOUD_APIKEY ...
	IBMCLOUD_APIKEY = "IBMCLOUD_APIKEY"
	// IBMCLOUD_PROFILEID ...
	IBMCLOUD_PROFILEID = "IBMCLOUD_PROFILEID"
	// IAM ...
	IAM = "iam"
	// PODIDENTITY ...
	PODIDENTITY = "pod-identity"
)