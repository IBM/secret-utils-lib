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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	b64 "encoding/base64"

	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	FakeAuthType string
)

// FakeGetk8sClientSet ...
func FakeGetk8sClientSet(logger *zap.Logger) (*fake.Clientset, error) {
	logger.Info("Getting fake k8s client")
	return fake.NewSimpleClientset(), nil
}

// FakeGetNameSpace ...
func FakeGetNameSpace(logger *zap.Logger) (string, error) {
	return "kube-system", nil
}

// FakeGetSecretData ...
func FakeGetSecretData(logger *zap.Logger) (string, string, error) {
	logger.Info("Fetching secret data")

	clientset, err := FakeGetk8sClientSet(logger)
	if err != nil {
		logger.Error("Error fetching k8s client set", zap.Error(err))
		return "", "", utils.Error{Description: utils.ErrFetchingSecrets, BackendError: err.Error()}
	}

	namespace, err := FakeGetNameSpace(logger)
	if err != nil {
		logger.Error("Unable to fetch namespace", zap.Error(err))
		return "", "", utils.Error{Description: utils.ErrFetchingSecretNoNamespace, BackendError: err.Error()}
	}

	err = FakeCreateSecret(logger, clientset)
	if err != nil {
		logger.Error("Unable to create secret", zap.Error(err))
		return "", "", utils.Error{Description: "Error creating fake secret", BackendError: err.Error()}
	}

	return FakeGetCredentials(logger, clientset, namespace)
}

// FakeCreateSecret ...
func FakeCreateSecret(logger *zap.Logger, clientset *fake.Clientset) error {
	secret := new(v1.Secret)

	var secretfilepath, dataname string
	switch FakeAuthType {
	case utils.IAM:
		secret.Name = utils.IBMCLOUD_CREDENTIALS_SECRET
		secretfilepath = "test-fixtures/ibmcloud_credentials/valid/apikey.toml"
		dataname = utils.CLOUD_PROVIDER_ENV
	case utils.PODIDENTITY:
		secret.Name = utils.IBMCLOUD_CREDENTIALS_SECRET
		secretfilepath = "test-fixtures/ibmcloud_credentials/valid/trusted_profile.toml"
		dataname = utils.CLOUD_PROVIDER_ENV
	case utils.DEFAULT:
		secret.Name = utils.STORAGE_SECRET_STORE_SECRET
		secretfilepath = "test-fixtures/valid/slclient.toml"
		dataname = utils.SECRET_STORE_FILE
	default:
		return errors.New("undefined auth type")
	}

	secret.Namespace = "kube-system"
	data := make(map[string][]byte)
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	configPath := filepath.Join(pwd, "..", "..", secretfilepath)

	byteData, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error("Error reading secret data", zap.Error(err))
		return err
	}

	dst := make([]byte, b64.StdEncoding.EncodedLen(len(byteData)))
	b64.StdEncoding.Encode(dst, byteData)
	data[dataname] = dst
	secret.Data = data
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return err
	}
	return nil
}

// FakeGetCredentials ...
func FakeGetCredentials(logger *zap.Logger, clientset *fake.Clientset, namespace string) (string, string, error) {
	logger.Info("Trying to fetch ibm-cloud-credentials secret")

	var dataname string
	var secretname string
	// Fetching secret
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), utils.IBMCLOUD_CREDENTIALS_SECRET, metav1.GetOptions{})
	if err == nil {
		dataname = utils.CLOUD_PROVIDER_ENV
		secretname = utils.IBMCLOUD_CREDENTIALS_SECRET
	} else {
		logger.Error("Unable to find secret", zap.Error(err), zap.String("Secret name", utils.IBMCLOUD_CREDENTIALS_SECRET))
		logger.Info("Trying to fetch storage-secret-store secret")
		secret, err = clientset.CoreV1().Secrets(namespace).Get(context.TODO(), utils.STORAGE_SECRET_STORE_SECRET, metav1.GetOptions{})
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
