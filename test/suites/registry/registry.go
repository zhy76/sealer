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

package registry

import (
	"fmt"
	"strings"

	"github.com/sealerio/sealer/test/testhelper"
	"github.com/sealerio/sealer/test/testhelper/settings"

	"github.com/onsi/gomega"
)

func RemoteLogin(sshClient *testhelper.SSHClient) {
	cmd := fmt.Sprintf("%s login %s -u %s -p %s", settings.DefaultSealerBin, settings.RegistryURL,
		settings.RegistryUsername,
		settings.RegistryPasswd)
	result, err := sshClient.SSH.CmdToString(sshClient.RemoteHostIP, nil, cmd, "")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	testhelper.CheckBeTrue(func() bool {
		return strings.Contains(result, "Login Succeeded!")
	}())
}

func RemoteLogout(sshClient *testhelper.SSHClient) {
	testhelper.DeleteFileRemotely(sshClient, DefaultRegistryAuthConfigDir())
}

// DefaultRegistryAuthConfigDir using root privilege to run sealer cmd at e2e test
func DefaultRegistryAuthConfigDir() string {
	return settings.DefaultRegistryAuthFileDir
}
