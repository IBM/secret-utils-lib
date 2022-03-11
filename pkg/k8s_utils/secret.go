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

const (
	// nameSpacePath is the path from which namespace where the pod is running is obtained.
	nameSpacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

// GetSecretData ...
func GetSecretData(logger *zap.Logger) (string, string, error) {
	logger.Info("Fetching secret")

	// Fetching cluster config used to create k8s client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("Error fetching in cluster config", zap.Error(err))
		return "", "", utils.Error{Description: utils.ErrFetchingSecretNoClusterConfig, BackendError: err.Error()}
	}

	// Creating k8s client used to read secret
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Error("Error creating k8s client", zap.Error(err))
		return "", "", utils.Error{Description: utils.ErrFetchingSecretNoK8sClient, BackendError: err.Error()}
	}

	// Reading the namespace in which the pod is deployed
	byteData, err := ioutil.ReadFile(nameSpacePath)
	if err != nil {
		logger.Error("Error fetching namespace", zap.Error(err))
		return "", "", utils.Error{Description: utils.ErrFetchingSecretNoNamespace, BackendError: err.Error()}
	}

	namespace := string(byteData)
	if namespace == "" {
		logger.Error("Unable to fetch namespace", zap.Error(err))
		return "", "", utils.Error{Description: utils.ErrFetchingSecretNoNamespace}
	}

	logger.Info("Trying to fetch ibm-cloud-credentials secret")

	var dataname string
	var secretname string
	// Fetching secret
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), utils.IBMCLOUD_CREDENTIALS_SECRET, v1.GetOptions{})
	if err == nil {
		dataname = utils.CLOUD_PROVIDER_ENV
		secretname = utils.IBMCLOUD_CREDENTIALS_SECRET
	} else {
		logger.Error("Unable to find secret", zap.Error(err), zap.String("Secret name", utils.IBMCLOUD_CREDENTIALS_SECRET))
		logger.Info("Trying to fetch storage-secret-store secret")
		secret, err = clientset.CoreV1().Secrets(namespace).Get(context.TODO(), utils.STORAGE_SECRET_STORE_SECRET, v1.GetOptions{})
		if err != nil {
			logger.Error("Unable to find secret", zap.Error(err), zap.String("Secret name", utils.STORAGE_SECRET_STORE_SECRET))
			return "", "", utils.Error{Description: utils.ErrFetchingSecrets, BackendError: err.Error()}
		}
		dataname = utils.SECRET_STORE_FILE
		secretname = utils.STORAGE_SECRET_STORE_SECRET
	}

	if secret.Data == nil {
		logger.Error("No data found in the secret")
		return "", "", utils.Error{Description: fmt.Sprintf(utils.ErrEmptyDataInSecret, secretname)}
	}

	byteData, ok := secret.Data[dataname]
	if !ok {
		logger.Error("Expected data not found in the secret")
		return "", "", utils.Error{Description: fmt.Sprintf(utils.ErrExpectedDataNotFound, dataname, secretname)}
	}

	sEnc := b64.StdEncoding.EncodeToString(byteData)

	sDec, err := b64.StdEncoding.DecodeString(sEnc)
	if err != nil {
		logger.Error("Error decoding the secret data", zap.Error(err), zap.String("Secret name", secretname), zap.String("Data name", dataname))
		return "", "", utils.Error{Description: fmt.Sprintf(utils.ErrFetchingSecretData, secretname, dataname), BackendError: err.Error()}
	}

	logger.Info("Successfully fetched secret data")
	return string(sDec), secretname, nil
}
