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

package config

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// clusterInfo ...
	clusterInfoCM = "cluster-info"
	// clusterConfigName ...
	clusterConfigName = "cluster-config.json"
	// stageMasterURLsubstr ...
	stageMasterURLsubstr = ".test."
	// tokenExchangePath ...
	tokenExchangePath = "/identity/token"
	// constTrue ...
	constTrue = "True"
	// maxNodes ...
	maxNodes = 3
)

// ClusterConfig ...
type ClusterConfig struct {
	ClusterID string `json:"cluster_id"`
	MasterURL string `json:"master_url"`
}

// GetClusterInfo ...
func GetClusterInfo(kc k8s_utils.KubernetesClient, logger *zap.Logger) (ClusterConfig, error) {
	data, err := k8s_utils.GetConfigMapData(kc, clusterInfoCM, clusterConfigName)
	var cc ClusterConfig
	if err != nil {
		logger.Error("Error fetching cluster info", zap.Error(err))
		return cc, err
	}

	err = json.Unmarshal([]byte(data), &cc)
	if err != nil {
		logger.Error("Error fetching cluster-info configmap", zap.Error(err))
		return cc, utils.Error{Description: utils.ErrFetchingClusterConfig, BackendError: err.Error()}
	}

	return cc, nil
}

// FrameTokenExchangeURL ...
func FrameTokenExchangeURL(kc k8s_utils.KubernetesClient, providerType string, logger *zap.Logger) string {

	// Fetch token exchange URL from cloud-conf
	cloudConf, err := GetCloudConf(logger, kc)
	if err == nil && cloudConf.TokenExchangeURL != "" {
		return cloudConf.TokenExchangeURL + tokenExchangePath
	}

	logger.Info("Unable to fetch token exchange URL from cloud-conf")
	secret, err := k8s_utils.GetSecretData(kc, utils.STORAGE_SECRET_STORE_SECRET, utils.SECRET_STORE_FILE)
	if err == nil {
		if secretConfig, err := ParseConfig(logger, secret); err == nil {
			url, err := GetTokenExchangeURLfromStorageSecretStore(kc, *secretConfig, providerType)
			if err == nil {
				return url
			}
		}
	}

	logger.Info("Unable to fetch token exchange URL using secret, forming url using cluster info")
	cc, err := GetClusterInfo(kc, logger)
	if err != nil {
		logger.Error("Error fetching cluster master URL", zap.Error(err))
		return (utils.ProdPublicIAMURL + tokenExchangePath)
	}

	return FrameTokenExchangeURLFromClusterInfo(kc, cc, logger)
}

// GetTokenExchangeURLfromStorageSecretStore ...
func GetTokenExchangeURLfromStorageSecretStore(kc k8s_utils.KubernetesClient, config Config, providerType string) (string, error) {

	var url string
	switch providerType {
	case utils.VPC:
		url = config.VPC.G2TokenExchangeURL
	case utils.Bluemix:
		url = config.Bluemix.IamURL
	case utils.Softlayer:
		url = config.Softlayer.SoftlayerTokenExchangeURL
	}

	if url == "" {
		return "", utils.Error{Description: utils.WarnFetchingTokenExchangeURL}
	}

	isSatellite, err := isSatellite(kc)
	if err != nil {
		return "", utils.Error{Description: "Unable to determine IAAS provider type", BackendError: err.Error()}
	}

	isProd := isProduction(url)

	// If the cluster is satellite, always use public IAM URL.
	if isSatellite {
		if isProd {
			return utils.ProdPublicIAMURL + tokenExchangePath, nil
		}
		return utils.StagePublicIAMURL + tokenExchangePath, nil
	}

	if isProd {
		return utils.ProdPrivateIAMURL + tokenExchangePath, nil
	}
	return utils.StagePrivateIAMURL + tokenExchangePath, nil
}

// FrameTokenExchangeURLFromClusterInfo ...
func FrameTokenExchangeURLFromClusterInfo(kc k8s_utils.KubernetesClient, cc ClusterConfig, logger *zap.Logger) string {
	isSatellite, _ := isSatellite(kc)

	if !strings.Contains(cc.MasterURL, stageMasterURLsubstr) {
		logger.Info("Env-Production")
		if isSatellite {
			return (utils.ProdPublicIAMURL + tokenExchangePath)
		}
		return (utils.ProdPrivateIAMURL + tokenExchangePath)
	}

	logger.Info("Env-Stage")
	if isSatellite {
		return (utils.StagePublicIAMURL + tokenExchangePath)
	}
	return (utils.StagePrivateIAMURL + tokenExchangePath)
}

// isProduction determines if the env in which a pod is deployed is stage or production
func isProduction(url string) bool {
	if !strings.Contains(url, "stage") && !strings.Contains(url, "test") {
		return true
	}
	return false
}

// isSatelliteCluster checks if the cluster where the pod is currently running is a satellite cluster or not
func isSatellite(kc k8s_utils.KubernetesClient) (bool, error) {
	nodeList, err := kc.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{Limit: maxNodes})
	if err != nil {
		return false, err
	}

	if len(nodeList.Items) == 0 {
		return false, utils.Error{Description: "No nodes found"}
	}

	for _, node := range nodeList.Items {
		providerType := node.ObjectMeta.Labels[utils.IaasProviderNodeLabel]
		if providerType == utils.SatelliteProvider {
			return true, nil
		}
	}

	return false, nil
}
