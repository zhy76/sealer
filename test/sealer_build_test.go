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
	"os"
	"path/filepath"

	"github.com/sealerio/sealer/test/suites/build"
	"github.com/sealerio/sealer/test/suites/image"
	"github.com/sealerio/sealer/test/suites/registry"
	"github.com/sealerio/sealer/test/testhelper"
	"github.com/sealerio/sealer/test/testhelper/client/docker"
	"github.com/sealerio/sealer/test/testhelper/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sealer build", func() {
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
		client.DeleteTestContainer(id)
	})

	Context("testing build with cmds", func() {
		BeforeEach(func() {
			buildPath := filepath.Join(build.WithCmdsBuildDir())
			Eventually(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, buildPath, buildPath)
				return err == nil
			}, settings.MaxWaiteTime).Should(BeTrue())
		})
		AfterEach(func() {
			err := os.Chdir(settings.DefaultTestEnvDir)
			testhelper.CheckErr(err)
		})

		It("start to build with cmds", func() {
			imageName := build.GetBuildImageName()
			cmd := fmt.Sprintf("cd %s && ", build.WithCmdsBuildDir()) +
				build.NewArgsOfBuild().
					SetKubeFile("Kubefile").
					SetImageName(imageName).
					SetContext(".").
					String()
			build.RemoteSealerBuild(sshClient, cmd)
			// check: sealer images whether image exist
			testhelper.CheckBeTrue(build.RemoteCheckIsImageExist(sshClient, imageName))

			//TODO check image spec content
			// 1. launch cmds
			// 2. containerImageList:
			//docker.io/library/nginx:alpine
			//docker.io/library/busybox:latest

			// clean: build image
			image.RemoteDoImageOps(sshClient, "rmi", imageName)
		})

	})

	Context("testing build with launch", func() {
		BeforeEach(func() {
			buildPath := filepath.Join(build.WithLaunchBuildDir())
			Eventually(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, buildPath, buildPath)
				return err == nil
			}, settings.MaxWaiteTime).Should(BeTrue())
		})
		AfterEach(func() {
			err := os.Chdir(settings.DefaultTestEnvDir)
			testhelper.CheckErr(err)
		})

		It("start to build with launch", func() {
			imageName := build.GetBuildImageName()
			cmd := fmt.Sprintf("cd %s && ", build.WithLaunchBuildDir()) +
				build.NewArgsOfBuild().
					SetKubeFile("Kubefile").
					SetImageName(imageName).
					SetContext(".").
					String()
			build.RemoteSealerBuild(sshClient, cmd)

			// check: sealer images whether image exist
			testhelper.CheckBeTrue(build.RemoteCheckIsImageExist(sshClient, imageName))

			//TODO check image spec content
			// 1. launch app names
			// 2. containerImageList:
			//docker.io/library/nginx:alpine
			//docker.io/library/busybox:latest

			// clean: build image
			image.RemoteDoImageOps(sshClient, "rmi", imageName)
		})

	})

	Context("testing build with app cmds", func() {
		BeforeEach(func() {
			buildPath := filepath.Join(build.WithAPPCmdsBuildDir())
			Eventually(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, buildPath, buildPath)
				return err == nil
			}, settings.MaxWaiteTime).Should(BeTrue())
		})
		AfterEach(func() {
			err := os.Chdir(settings.DefaultTestEnvDir)
			testhelper.CheckErr(err)
		})

		It("start to build with app cmds", func() {
			imageName := build.GetBuildImageName()
			cmd := fmt.Sprintf("cd %s && ", build.WithAPPCmdsBuildDir()) +
				build.NewArgsOfBuild().
					SetKubeFile("Kubefile").
					SetImageName(imageName).
					SetContext(".").
					String()
			build.RemoteSealerBuild(sshClient, cmd)

			// check: sealer images whether image exist
			testhelper.CheckBeTrue(build.RemoteCheckIsImageExist(sshClient, imageName))

			//TODO check image spec content
			// 1. launch app names
			// 2. launch app cmds:

			// clean: build image
			image.RemoteDoImageOps(sshClient, "rmi", imageName)
		})

	})

	Context("testing build with --image-list flag", func() {
		BeforeEach(func() {
			buildPath := filepath.Join(build.WithImageListFlagBuildDir())
			Eventually(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, buildPath, buildPath)
				return err == nil
			}, settings.MaxWaiteTime).Should(BeTrue())
		})
		AfterEach(func() {
			err := os.Chdir(settings.DefaultTestEnvDir)
			testhelper.CheckErr(err)
		})

		It("start to build with --image-list flag", func() {
			imageName := build.GetBuildImageName()
			cmd := fmt.Sprintf("cd %s && ", build.WithImageListFlagBuildDir()) +
				build.NewArgsOfBuild().
					SetKubeFile("Kubefile").
					SetImageName(imageName).
					SetImageList("imagelist").
					SetContext(".").
					String()
			build.RemoteSealerBuild(sshClient, cmd)

			// check: sealer images whether image exist
			testhelper.CheckBeTrue(build.RemoteCheckIsImageExist(sshClient, imageName))

			//TODO check image spec content
			// 2. containerImageList:
			//docker.io/library/nginx:alpine
			//docker.io/library/busybox:latest

			// clean: build image
			image.RemoteDoImageOps(sshClient, "rmi", imageName)
		})

	})

	Context("testing multi platform build scenario", func() {
		BeforeEach(func() {
			buildPath := filepath.Join(build.WithMultiArchBuildDir())
			Eventually(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, buildPath, buildPath)
				return err == nil
			}, settings.MaxWaiteTime).Should(BeTrue())
		})
		AfterEach(func() {
			registry.RemoteLogout(sshClient)
			err := os.Chdir(settings.DefaultTestEnvDir)
			testhelper.CheckErr(err)
		})

		It("multi build only with amd64", func() {
			imageName := build.GetBuildImageName()
			cmd := fmt.Sprintf("cd %s && ", build.WithMultiArchBuildDir()) +
				build.NewArgsOfBuild().
					SetKubeFile("Kubefile").
					SetImageName(imageName).
					SetPlatforms([]string{"linux/amd64"}).
					SetContext(".").
					String()
			build.RemoteSealerBuild(sshClient, cmd)

			// check: sealer images whether image exist
			testhelper.CheckBeTrue(build.RemoteCheckIsImageExist(sshClient, imageName))

			// check: push build image
			image.RemoteDoImageOps(sshClient, "push", imageName)

			// clean: build image
			image.RemoteDoImageOps(sshClient, "rmi", imageName)

		})

		It("multi build only with arm64", func() {
			imageName := build.GetBuildImageName()
			cmd := build.NewArgsOfBuild().
				SetKubeFile("Kubefile").
				SetImageName(imageName).
				SetPlatforms([]string{"linux/arm64"}).
				SetContext(".").
				String()
			build.RemoteSealerBuild(sshClient, cmd)
			// check: sealer images whether image exist
			testhelper.CheckBeTrue(build.CheckIsMultiArchImageExist(imageName))

			// check: push build image
			image.RemoteDoImageOps(sshClient, "push", imageName)

			// clean: build image
			image.RemoteDoImageOps(sshClient, "rmi", imageName)
		})

		It("multi build with amd64 and arm64", func() {
			imageName := build.GetBuildImageName()
			cmd := build.NewArgsOfBuild().
				SetKubeFile("Kubefile").
				SetImageName(imageName).
				SetPlatforms([]string{"linux/amd64", "linux/arm64"}).
				SetContext(".").
				String()
			build.RemoteSealerBuild(sshClient, cmd)
			// check: sealer images whether image exist
			testhelper.CheckBeTrue(build.CheckIsMultiArchImageExist(imageName))

			// check: push build image
			image.RemoteDoImageOps(sshClient, "push", imageName)

			// clean: build image
			image.RemoteDoImageOps(sshClient, "rmi", imageName)
		})

	})

})
