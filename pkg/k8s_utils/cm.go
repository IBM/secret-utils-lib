package k8s_utils

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// clusterInfo contains the cluster information
type clusterInfo struct {
	ClusterID string `json:"cluster_id"`
	MasterURL string `json:"master_url"`
}

func GetCM(kc *KubernetesClient) {
	clientset := kc.clientset
	cm, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "cluster-info", metav1.GetOptions{})

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cm.Data)
	data := cm.Data["cluster-config.json"]
	cf := new(clusterInfo)
	err = json.Unmarshal([]byte(data), cf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cf)
}
