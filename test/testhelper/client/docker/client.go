// Copyright © 2023 Alibaba Group Holding Ltd.
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

package docker

import (
	"fmt"
	"net"

	"github.com/sealerio/sealer/pkg/infra/container"
	dc "github.com/sealerio/sealer/pkg/infra/container/client"
	"github.com/sealerio/sealer/test/testhelper"
	"github.com/sirupsen/logrus"
)

type Client struct {
	Provider dc.ProviderService
}

// Set up a docker client
func NewDockerClient() (*Client, error) {
	client, err := container.NewClientWithCluster(nil)

	if err != nil {
		return nil, fmt.Errorf("new docker client failed: %v", err)
	}
	return &Client{
		Provider: client.Provider,
	}, nil
}

// Create a container for test env and return container id and container networkname
func (c *Client) CreateTestContainer() (string, string, error) {
	id, err := c.Provider.RunContainer(&dc.CreateOptsForContainer{
		ContainerName:     "test-container",
		ContainerHostName: "test-container-host-name",
		ImageName:         container.ImageName,
		NetworkName:       container.NetworkName,
	})
	if err != nil {
		return id, container.NetworkName, fmt.Errorf("failed to run container: %v", err)
	}

	info, err := c.Provider.GetContainerInfo(id, container.NetworkName)
	if err != nil {
		return id, container.NetworkName, fmt.Errorf("failed to get container info of %s ,error is %v", id, err)
	}
	if info.Status != "running" {
		return id, container.NetworkName, fmt.Errorf("failed to get container info %s,container is %v", id, info.Status)
	}
	logrus.Infof("succuss to apply docker container")
	return id, container.NetworkName, nil
}

// Delete a container
func (c *Client) DeleteTestContainer(id string) {
	err := c.Provider.RmContainer(id)
	testhelper.CheckErr(err)
}

// Get a container ip
func (c *Client) GetContainerIP(id string, networkName string) net.IP {
	container, err := c.Provider.GetContainerInfo(id, networkName)
	testhelper.CheckErr(err)
	return net.ParseIP(container.ContainerIP)
}
