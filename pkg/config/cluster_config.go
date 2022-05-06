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
func FrameTokenExchangeURL(kc k8s_utils.KubernetesClient, logger *zap.Logger) (string, error) {
	logger.Info("Forming token exchange URL")
	cc, err := GetClusterInfo(kc, logger)
	if err != nil {
		// If the cluster-info is not found (this is the case of an unmanaged cluster, )
		if strings.Contains(err.Error(), "not found") {
			return utils.PublicTokenExchangeURL, nil
		}
		logger.Error("Error fetching cluster master URL", zap.Error(err))
		return "", err
	}

	if cc.MasterURL == "" {
		logger.Error("Empty cluster master url")
		return "", utils.Error{Description: utils.ErrFetchingClusterConfig, BackendError: "Empty cluster master URL"}
	}

	if !strings.Contains(cc.MasterURL, stageMasterURLsubstr) {
		logger.Info("Env - Production")
		return utils.ProdTokenExchangeURL, nil
	}

	logger.Info("Env - Stage")
	return utils.StageTokenExchangeURL, nil
}