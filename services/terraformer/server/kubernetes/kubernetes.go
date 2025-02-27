package kubernetes

import (
	"fmt"

	"github.com/Berops/claudie/proto/pb"
	"github.com/Berops/claudie/services/terraformer/server/clusterBuilder"
)

type K8Scluster struct {
	DesiredK8s  *pb.K8Scluster
	CurrentK8s  *pb.K8Scluster
	ProjectName string
}

func (k K8Scluster) Build() error {
	var currentInfo *pb.ClusterInfo
	// check if current cluster was defined, to avoid access of unrefferenced memory
	if k.CurrentK8s != nil {
		currentInfo = k.CurrentK8s.ClusterInfo
	}
	cluster := clusterBuilder.ClusterBuilder{
		DesiredInfo: k.DesiredK8s.ClusterInfo,
		CurrentInfo: currentInfo,
		ProjectName: k.ProjectName,
		ClusterType: pb.ClusterType_K8s}
	err := cluster.CreateNodepools()
	if err != nil {
		return fmt.Errorf("error while creating the K8s cluster %s : %v", k.DesiredK8s.ClusterInfo.Name, err)
	}
	return nil
}

func (k K8Scluster) Destroy() error {
	cluster := clusterBuilder.ClusterBuilder{
		//DesiredInfo: , //desired state is not used in DestroyNodepools
		CurrentInfo: k.CurrentK8s.ClusterInfo,
		ProjectName: k.ProjectName,
		ClusterType: pb.ClusterType_K8s}
	err := cluster.DestroyNodepools()
	if err != nil {
		return fmt.Errorf("error while destroying the K8s cluster %s : %v", k.CurrentK8s.ClusterInfo.Name, err)
	}
	return nil
}
