// Copyright © 2021 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sealerio/sealer/common"
	"github.com/sealerio/sealer/pkg/client/k8s"
	"github.com/sealerio/sealer/pkg/clusterfile"
	v12 "github.com/sealerio/sealer/pkg/define/image/v1"
	"github.com/sealerio/sealer/pkg/define/options"
	"github.com/sealerio/sealer/pkg/imageengine"
	"github.com/sealerio/sealer/pkg/infradriver"
	v2 "github.com/sealerio/sealer/types/api/v2"
	"github.com/sealerio/sealer/utils/strings"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var applyClusterFile string

var applyMode string

const MasterRoleLabel = "node-role.kubernetes.io/master"

var longApplyCmdDescription = `apply command is used to apply a Kubernetes cluster via specified Clusterfile.
If the Clusterfile is applied first time, Kubernetes cluster will be created. Otherwise, sealer
will apply the diff change of current Clusterfile and the original one.`

var exampleForApplyCmd = `
  sealer apply -f Clusterfile
`

func NewApplyCmd() *cobra.Command {
	applyCmd := &cobra.Command{
		Use:     "apply",
		Short:   "apply a Kubernetes cluster via specified Clusterfile",
		Long:    longApplyCmdDescription,
		Example: exampleForApplyCmd,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				cf              clusterfile.Interface
				clusterFileData []byte
				err             error
			)
			logrus.Warn("sealer apply command will be deprecated in the future, please use sealer run instead.")

			if applyClusterFile == "" {
				return fmt.Errorf("you must input Clusterfile")
			}

			clusterFileData, err = os.ReadFile(filepath.Clean(applyClusterFile))
			if err != nil {
				return err
			}

			cf, err = clusterfile.NewClusterFile(clusterFileData)
			if err != nil {
				return err
			}

			desiredCluster := cf.GetCluster()
			infraDriver, err := infradriver.NewInfraDriver(&desiredCluster)
			if err != nil {
				return err
			}

			// use image extension to determine apply type:
			// scale up cluster, install applications, maybe support upgrade later
			imageName := desiredCluster.Spec.Image
			imageEngine, err := imageengine.NewImageEngine(options.EngineGlobalConfigurations{})
			if err != nil {
				return err
			}

			extension, err := imageEngine.GetSealerImageExtension(&options.GetImageAnnoOptions{ImageNameOrID: imageName})
			if err != nil {
				return fmt.Errorf("failed to get cluster image extension: %s", err)
			}

			if extension.Type == v12.AppInstaller {
				logrus.Infof("start to install application: %s", imageName)
				return installApplication(imageName, []string{}, extension, infraDriver, imageEngine)
			}

			client := getClusterClient()
			if client == nil {
				// no k8s client means to init a new cluster.
				logrus.Infof("start to create new cluster with image: %s", imageName)
				return createNewCluster(imageName, infraDriver, imageEngine, cf)
			}

			currentCluster, err := GetCurrentCluster(client)
			if err != nil {
				return errors.Wrap(err, "failed to get current cluster")
			}

			mj, md := strings.Diff(currentCluster.GetMasterIPList(), desiredCluster.GetMasterIPList())
			nj, nd := strings.Diff(currentCluster.GetNodeIPList(), desiredCluster.GetNodeIPList())
			if len(mj) == 0 && len(md) == 0 && len(nj) == 0 && len(nd) == 0 {
				return nil
			}

			if len(md) > 0 || len(nd) > 0 {
				return fmt.Errorf("scale down not supported: %v,%v", md, nd)
			}
			logrus.Infof("start to scale up cluster with image: %s", imageName)
			return scaleUpCluster(imageName, mj, nj, infraDriver, imageEngine, cf)
		},
	}
	applyCmd.Flags().BoolVar(&ForceDelete, "force", false, "force to delete the specified cluster if set true")
	applyCmd.Flags().StringVarP(&applyClusterFile, "Clusterfile", "f", "", "Clusterfile path to apply a Kubernetes cluster")
	applyCmd.Flags().StringVarP(&applyMode, "applyMode", "m", common.ApplyModeApply, "load images to the specified registry in advance")
	return applyCmd
}

func GetCurrentCluster(client *k8s.Client) (*v2.Cluster, error) {
	nodes, err := client.ListNodes()
	if err != nil {
		return nil, err
	}

	cluster := &v2.Cluster{}
	var masterIPList []net.IP
	var nodeIPList []net.IP

	for _, node := range nodes.Items {
		addr := getNodeAddress(node)
		if addr == nil {
			continue
		}
		if _, ok := node.Labels[MasterRoleLabel]; ok {
			masterIPList = append(masterIPList, addr)
			continue
		}
		nodeIPList = append(nodeIPList, addr)
	}
	cluster.Spec.Hosts = []v2.Host{{IPS: masterIPList, Roles: []string{common.MASTER}}, {IPS: nodeIPList, Roles: []string{common.NODE}}}

	return cluster, nil
}

func getNodeAddress(node corev1.Node) net.IP {
	if len(node.Status.Addresses) < 1 {
		return nil
	}
	return net.ParseIP(node.Status.Addresses[0].Address)
}

func getClusterClient() *k8s.Client {
	client, err := k8s.NewK8sClient()
	if client != nil {
		return client
	}
	if err != nil {
		logrus.Warnf("try to new k8s client via default kubeconfig, maybe this is a new cluster that needs to be created: %v", err)
	}
	return nil
}
