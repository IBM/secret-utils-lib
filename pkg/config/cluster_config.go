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
	"encoding/json"
	"os"
	"strings"

	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
)

const (
	// clusterInfo ...
	clusterInfoCM = "cluster-info"
	// clusterConfigName ...
	clusterConfigName = "cluster-config.json"
	// stageMasterURLsubstr ...
	stageMasterURLsubstr = ".test."
	// satellite ...
	satellite clusterType = "satellite"
	// ipi ...
	ipi clusterType = "ipi"
	// managed ...
	managed clusterType = "managed"
	// tokenExchangePath ...
	tokenExchangePath = "/identity/token"
)

// clusterType refers to type of a cluster
type clusterType string

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

// getClusterType ...
func getClusterType() clusterType {
	if os.Getenv("IS_SATELLITE") == "True" {
		return satellite
	}

	if iksEnabled := os.Getenv("IKS_ENABLED"); strings.ToLower(iksEnabled) == "true" {
		return managed
	}

	return ipi
}

// getTokenExchangeURLfromSecret ...
func getTokenExchangeURLfromSecret(secret string, logger *zap.Logger) (string, error) {
	logger.Info("Framing token exchange URL using storage-secret-store")

	secretConfig, err := ParseConfig(logger, secret)
	if err != nil {
		return "", err
	}

	clustertype := getClusterType()
	logger.Info("Fetched cluster type", zap.String("ClusterType", string(clustertype)))

	var url string
	switch clustertype {
	case satellite:
		// Using provided url for token exchange if cluster type = satellite
		url = secretConfig.VPC.G2TokenExchangeURL
	case ipi:
		url = utils.ProdIAMURL
	case managed:
		if !strings.Contains(secretConfig.VPC.G2TokenExchangeURL, "stage") {
			url = utils.ProdIAMURL
		} else {
			url = utils.StageIAMURL
		}
	}

	if url == "" {
		return "", utils.Error{Description: utils.WarnFetchingTokenExchangeURL}
	}
	// Appending the base URL and token exchange path
	url = url + tokenExchangePath

	return url, nil
}

// FrameTokenExchangeURL ...
func FrameTokenExchangeURL(kc k8s_utils.KubernetesClient, logger *zap.Logger) (string, error) {
	logger.Info("Forming token exchange URL")

	secret, err := k8s_utils.GetSecret(kc, utils.STORAGE_SECRET_STORE_SECRET, utils.SECRET_STORE_FILE)
	if err == nil {
		url, err := getTokenExchangeURLfromSecret(secret, logger)
		if err == nil {
			return url, nil
		}
	}

	logger.Info("Unable to fetch token exchange URL using secret, forming url using cluster info")
	cc, err := GetClusterInfo(kc, logger)
	if err != nil {
		logger.Error("Error fetching cluster master URL", zap.Error(err))
		return (utils.PublicIAMURL + tokenExchangePath), nil
	}

	if !strings.Contains(cc.MasterURL, stageMasterURLsubstr) {
		logger.Info("Env - Production")
		return (utils.ProdIAMURL + tokenExchangePath), nil
	}

	logger.Info("Env - Stage")
	return (utils.StageIAMURL + tokenExchangePath), nil
}
