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

// Package k8s_utils ...
package k8s_utils

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetSecretData ...
func GetSecretData(secretname, dataname string, logger *zap.Logger) (string, error) {
	logger.Info("Fetching secret", zap.String("Secret name", secretname))

	// Fetching cluster config used to create k8s client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("Error fetching in cluster config", zap.Error(err))
		return "", utils.Error{Description: fmt.Sprintf("Error fetching secret %s - unable to fetch cluster config", secretname), BackendError: err.Error()}
	}

	// Creating k8s client used to read secret
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Error("Error creating k8s client", zap.Error(err))
		return "", utils.Error{Description: fmt.Sprintf("Error fetching secret %s - unable to create k8s client", secretname), BackendError: err.Error()}
	}

	// Reading the namespace in which the pod is deployed
	byteData, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		logger.Error("Error fetching namespace", zap.Error(err))
		return "", utils.Error{Description: fmt.Sprintf("Error fetching secret %s - unable to read namespace", secretname), BackendError: err.Error()}
	}

	namespace := string(byteData)
	if namespace == "" {
		logger.Error("Unable to fetch namespace", zap.Error(err))
		return "", utils.Error{Description: fmt.Sprintf("Error fetching secret %s - unable to fetch namespace", secretname)}
	}

	// Fetching secret
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretname, v1.GetOptions{})
	if err != nil {
		logger.Error("Unable to find secret", zap.Error(err), zap.String("Secret name", secretname))
		return "", utils.Error{Description: fmt.Sprintf("Error fetching secret %s - unable to fetch namespace", secretname), BackendError: err.Error()}
	}

	if secret.Data == nil {
		logger.Error("No data found in the secret")
		return "", utils.Error{Description: fmt.Sprintf("No data found in the secrer %s", secretname)}
	}

	byteData, ok := secret.Data[dataname]
	if !ok {
		logger.Error("Expected data not found in the secret")
		return "", utils.Error{Description: fmt.Sprintf("Expected data %s not found in the secret %s", dataname, secretname)}
	}

	sEnc := b64.StdEncoding.EncodeToString(byteData)

	sDec, err := b64.StdEncoding.DecodeString(sEnc)
	if err != nil {
		logger.Error("Error decoding the secret data", zap.Error(err), zap.String("Secret name", secretname), zap.String("Data name", dataname))
		return "", utils.Error{Description: fmt.Sprintf("Unable to fetch data from secret, Secret: %s, Data: %s", secretname, dataname), BackendError: err.Error()}
	}

	logger.Info("Successfully fetched secret data")
	return string(sDec), nil
}
