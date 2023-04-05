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

package test

import (
	"fmt"

	"github.com/sealerio/sealer/test/suites/build"
	"github.com/sealerio/sealer/test/suites/image"
	"github.com/sealerio/sealer/test/suites/registry"
	"github.com/sealerio/sealer/test/testhelper"
	"github.com/sealerio/sealer/test/testhelper/client/docker"
	"github.com/sealerio/sealer/test/testhelper/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sealer image", func() {
	Context("test sealer image on container", func() {
		It("test pull,tag,remove,push image", func() {
			By("start to prepare infra")
			client, err := docker.NewDockerClient()
			testhelper.CheckErr(err)
			id, networkName, err := client.CreateTestContainer()
			testhelper.CheckErr(err)
			defer client.DeleteTestContainer(id)
			ip := client.GetContainerIP(id, networkName)
			sshClient := testhelper.NewSSHClientByIP(ip)
			testhelper.CheckFuncBeTrue(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, settings.DefaultSealerBin, settings.DefaultSealerBin)
				return err == nil
			}, settings.MaxWaiteTime)

			By(fmt.Sprintf("pull image %s", settings.TestImageName))
			image.RemoteDoImageOps(sshClient, "pull", settings.TestImageName)
			testhelper.CheckBeTrue(build.RemoteCheckIsImageExist(sshClient, settings.TestImageName))

			faultImageNames := []string{
				fmt.Sprintf("%s/%s:latest", settings.DefaultImageName, settings.DefaultImageRepo),
				fmt.Sprintf("%s:latest", settings.DefaultImageDomain),
				fmt.Sprintf("%s:latest", settings.DefaultImageRepo),
			}

			for _, faultImageName := range faultImageNames {
				faultImageName := faultImageName
				By(fmt.Sprintf("pull fault image %s", faultImageName), func() {
					testhelper.CheckBeTrue(func() bool {
						err := sshClient.SSH.CmdAsync(sshClient.RemoteHostIP, nil, fmt.Sprintf("%s pull %s", settings.DefaultSealerBin, faultImageName))
						return err != nil
					}())

					testhelper.CheckNotBeTrue(build.RemoteCheckIsImageExist(sshClient, faultImageName))
				})
			}

			tagImageNames := []string{
				"e2eimage_test:latest",
				"e2eimage_test:v0.0.1",
				"sealer-io/e2eimage_test:v0.0.2",
				"docker.io/sealerio/e2eimage_test:v0.0.3",
			}

			By("tag by image name", func() {
				for _, newOne := range tagImageNames {
					image.RemoteTagImages(sshClient, settings.TestImageName, newOne)
					Expect(build.RemoteCheckIsImageExist(sshClient, newOne)).Should(BeTrue())
				}

				image.RemoteDoImageOps(sshClient, "images", "")

				for _, imageName := range tagImageNames {
					removeImage := imageName
					image.RemoteDoImageOps(sshClient, "rmi", removeImage)
				}

			})

			By("remove tag image", func() {
				tagImageName := "e2e_images_test:v0.3"
				image.RemoteDoImageOps(sshClient, "pull", settings.TestImageName)
				image.RemoteTagImages(sshClient, settings.TestImageName, tagImageName)
				testhelper.CheckBeTrue(build.RemoteCheckIsImageExist(sshClient, tagImageName))
				image.RemoteDoImageOps(sshClient, "rmi", tagImageName)
				testhelper.CheckNotBeTrue(build.RemoteCheckIsImageExist(sshClient, tagImageName))
			})

			By("push image", func() {
				registry.RemoteLogin(sshClient)
				defer registry.RemoteLogout(sshClient)
				image.RemoteDoImageOps(sshClient, "pull", settings.TestImageName)
				pushImageName := "docker.io/sealerio/e2eimage_test:v0.0.1"
				if settings.RegistryURL != "" && settings.RegistryUsername != "" && settings.RegistryPasswd != "" {
					pushImageName = settings.RegistryURL + "/" + settings.RegistryUsername + "/" + "e2eimage_test:v0.0.1"
				}
				image.RemoteTagImages(sshClient, settings.TestImageName, pushImageName)
				image.RemoteDoImageOps(sshClient, "push", pushImageName)
			})

			By(fmt.Sprintf("remove image %s", settings.TestImageName), func() {
				image.RemoteDoImageOps(sshClient, "images", "")
				image.RemoteDoImageOps(sshClient, "pull", settings.TestImageName)
				testhelper.CheckBeTrue(build.RemoteCheckIsImageExist(sshClient, settings.TestImageName))
				image.RemoteDoImageOps(sshClient, "rmi", settings.TestImageName)
				testhelper.CheckNotBeTrue(build.RemoteCheckIsImageExist(sshClient, settings.TestImageName))
			})
		})
	})

	Context("login registry", func() {
		var (
			client      *docker.Client
			sshClient   *testhelper.SSHClient
			id          string
			networkName string
		)

		BeforeEach(func() {
			By("start to prepare infra")
			var err error
			client, err = docker.NewDockerClient()
			testhelper.CheckErr(err)
			id, networkName, err = client.CreateTestContainer()
			testhelper.CheckErr(err)

			ip := client.GetContainerIP(id, networkName)
			sshClient = testhelper.NewSSHClientByIP(ip)
			testhelper.CheckFuncBeTrue(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, settings.DefaultSealerBin, settings.DefaultSealerBin)
				return err == nil
			}, settings.MaxWaiteTime)
		})

		AfterEach(func() {
			registry.RemoteLogout(sshClient)
			client.DeleteTestContainer(id)
		})

		It("with correct name and password", func() {
			image.CheckLoginResult(
				settings.RegistryURL,
				settings.RegistryUsername,
				settings.RegistryPasswd,
				true)
		})
		It("with incorrect name and password", func() {
			image.CheckLoginResult(
				settings.RegistryURL,
				settings.RegistryPasswd,
				settings.RegistryUsername,
				false)
		})
		It("with only name", func() {
			image.CheckLoginResult(
				settings.RegistryURL,
				settings.RegistryUsername,
				"",
				false)
		})
		It("with only password", func() {
			image.CheckLoginResult(
				settings.RegistryURL,
				"",
				settings.RegistryPasswd,
				false)
		})
		It("with only registryURL", func() {
			image.CheckLoginResult(
				settings.RegistryURL,
				"",
				"",
				false)
		})
	})

	//todo add mount and umount e2e test
})
