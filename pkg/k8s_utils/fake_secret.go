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
	"os"
	"path/filepath"

	b64 "encoding/base64"

	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// FakeGetk8sClientSet ...
func FakeGetk8sClientSet(logger *zap.Logger) (*KubernetesClient, error) {
	logger.Info("Getting fake k8s client")
	return &KubernetesClient{namespace: "kube-system", logger: logger, clientset: fake.NewSimpleClientset()}, nil
}

// FakeCreateSecret ...
func FakeCreateSecret(kc *KubernetesClient, fakeAuthType string) error {
	secret := new(v1.Secret)

	var secretfilepath, dataname string
	switch fakeAuthType {
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

	secret.Namespace = kc.GetNameSpace()
	data := make(map[string][]byte)
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	configPath := filepath.Join(pwd, secretfilepath)
	byteData, err := os.ReadFile(configPath)
	if err != nil {
		kc.logger.Error("Error reading secret data", zap.Error(err))
		return err
	}

	dst := make([]byte, b64.StdEncoding.EncodedLen(len(byteData)))
	b64.StdEncoding.Encode(dst, byteData)
	data[dataname] = dst
	secret.Data = data
	clientset := kc.clientset
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		kc.logger.Error("Error creating secret", zap.Error(err))
		return err
	}
	return nil
}