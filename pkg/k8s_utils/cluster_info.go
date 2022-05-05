package k8s_utils

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// clusterInfo ...
	clusterInfoCM = "cluster-info"
	// clusterConfigName ...
	clusterConfigName = "cluster-config.json"
)

// clusterInfo contains the cluster information
type clusterInfo struct {
	MasterURL string `json:"master_url"`
}

// FrameTokenExchangeURL ...
func FrameTokenExchangeURL(kc *KubernetesClient) (string, error) {
	kc.logger.Info("Forming token exchange URL")
	masterUrl, err := getClusterMasterURL(kc)
	if err != nil {
		kc.logger.Error("Error fetching cluster master URL", zap.Error(err))
		return "", err
	}

	if !strings.Contains(masterUrl, ".test.") {
		kc.logger.Info("Env - Production")
		return utils.ProdIAMURL, nil
	}

	kc.logger.Info("Env - Stage")
	return utils.StageIAMURL, nil
}

// getClusterMasterURL ...
func getClusterMasterURL(kc *KubernetesClient) (string, error) {
	kc.logger.Info("Fetching cluster master URL")

	clientset := kc.clientset
	namespace := kc.GetNameSpace()

	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), clusterInfoCM, metav1.GetOptions{})

	if err != nil {
		kc.logger.Error("Error fetching cluster-info configmap", zap.Error(err))
		return "", err
	}

	data, ok := cm.Data[clusterConfigName]
	if !ok {
		kc.logger.Error("cluster-config.json is not present")
		return "", utils.Error{Description: utils.ErrEmptyClusterConfig}
	}

	cf := new(clusterInfo)
	err = json.Unmarshal([]byte(data), cf)
	if err != nil {
		kc.logger.Error("Error fetching cluster-info configmap", zap.Error(err))
		return "", utils.Error{Description: utils.ErrFetchingClusterConfig, BackendError: err.Error()}
	}

	if cf.MasterURL == "" {
		kc.logger.Error("Empty cluster master url")
		return "", utils.Error{Description: utils.ErrFetchingClusterConfig, BackendError: "Empty cluster master URL"}
	}

	return cf.MasterURL, nil
}
