// Copyright 2017 Authors of Cilium
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

package RuntimeTest

import (
	"context"
	"fmt"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

var _ = Describe("RuntimeMonitorTest", func() {

	var initialized bool
	var networkName string = "cilium-net"
	var logger *log.Entry
	var docker *helpers.Docker
	var cilium *helpers.Cilium

	initialize := func() {
		if initialized == true {
			return
		}
		logger = log.WithFields(log.Fields{"testName": "RuntimeMonitorTest"})
		logger.Info("Starting")
		docker, cilium = helpers.CreateNewRuntimeHelper("runtime", logger)
		cilium.WaitUntilReady(100)
		docker.NetworkCreate(networkName, "")

		res := cilium.PolicyEnforcementSet("default", false)
		Expect(res.WasSuccessful()).Should(BeTrue())

		initialized = true
	}

	BeforeEach(func() {
		initialize()
	})

	AfterEach(func() {
		docker.SampleContainersActions("delete", networkName)
	})

	It("Cilium monitor verbose mode", func() {

		res := cilium.Exec("config Debug=true DropNotification=true TraceNotification=true")
		Expect(res.WasSuccessful()).Should(BeTrue())

		ctx, cancel := context.WithCancel(context.Background())

		res = docker.Node.ExecContext(ctx, "sudo cilium monitor -v")
		docker.SampleContainersActions("create", networkName)
		helpers.Sleep(5)
		cancel()
		endpoints, err := cilium.GetEndpointsIds()
		Expect(err).Should(BeNil())

		for k, v := range endpoints {
			filter := fmt.Sprintf("FROM %s DEBUG:", v)
			docker.ContainerExec(k, "ping -c 1 httpd1")
			Expect(res.Output().String()).Should(ContainSubstring(filter))
		}
	})

	It("Cilium monitor event types", func() {
		eventTypes := map[string]string{
			"drop":    "DROP:",
			"debug":   "DEBUG:",
			"capture": "DEBUG:",
		}

		res := cilium.Exec("config Debug=true DropNotification=true TraceNotification=true")
		Expect(res.WasSuccessful()).Should(BeTrue())
		for k, v := range eventTypes {
			By(fmt.Sprintf("Type %s", k))

			ctx, cancel := context.WithCancel(context.Background())
			res := docker.Node.ExecContext(ctx, fmt.Sprintf("sudo cilium monitor --type %s -v", k))
			docker.SampleContainersActions("create", networkName)
			docker.ContainerExec("app1", "ping -c 4 httpd1")
			helpers.Sleep(5)
			cancel()

			Expect(res.CountLines()).Should(BeNumerically(">", 3))
			Expect(res.Output().String()).Should(ContainSubstring(v))
			docker.SampleContainersActions("delete", networkName)
		}
	})

	It("cilium monitor check --from", func() {
		res := cilium.Exec("config Debug=true DropNotification=true TraceNotification=true")
		Expect(res.WasSuccessful()).Should(BeTrue())

		docker.SampleContainersActions("create", networkName)
		endpoints, err := cilium.GetEndpointsIds()
		Expect(err).Should(BeNil())

		ctx, cancel := context.WithCancel(context.Background())
		res = docker.Node.ExecContext(ctx, fmt.Sprintf(
			"sudo cilium monitor --type debug --from %s -v", endpoints["app1"]))
		docker.ContainerExec("app1", "ping -c 5 httpd1")
		helpers.Sleep(5)
		cancel()

		Expect(res.CountLines()).Should(BeNumerically(">", 3))
		filter := fmt.Sprintf("FROM %s DEBUG:", endpoints["app1"])
		Expect(res.Output().String()).Should(ContainSubstring(filter))

		//Debug mode shouldn't have DROP lines
		Expect(res.Output().String()).ShouldNot(ContainSubstring("DROP"))

	})

	It("cilium monitor check --to", func() {

		res := cilium.Exec(
			"config Debug=true DropNotification=true TraceNotification=true PolicyEnforcement=always")
		Expect(res.WasSuccessful()).Should(BeTrue())

		docker.SampleContainersActions("create", networkName)
		endpoints, err := cilium.GetEndpointsIds()
		Expect(err).Should(BeNil())

		ctx, cancel := context.WithCancel(context.Background())
		res = docker.Node.ExecContext(ctx, fmt.Sprintf(
			"sudo cilium monitor --type drop -v --to %s", endpoints["httpd1"]))

		docker.ContainerExec("app1", "ping -c 5 httpd1")
		helpers.Sleep(5)
		cancel()

		Expect(res.CountLines()).Should(BeNumerically(">", 3))
		filter := fmt.Sprintf("FROM %s DROP:", endpoints["httpd1"])
		Expect(res.Output().String()).Should(ContainSubstring(filter))

	})

	It("cilium monitor check --related-to", func() {

		res := cilium.Exec(
			"config Debug=true DropNotification=true TraceNotification=true PolicyEnforcement=always")
		Expect(res.WasSuccessful()).Should(BeTrue())

		docker.SampleContainersActions("create", networkName)
		endpoints, err := cilium.GetEndpointsIds()
		Expect(err).Should(BeNil())

		ctx, cancel := context.WithCancel(context.Background())
		res = docker.Node.ExecContext(ctx, fmt.Sprintf(
			"sudo cilium monitor -v --type drop --related-to %s", endpoints["httpd1"]))
		cilium.EndpointWaitUntilReady()
		docker.ContainerExec("app1", "curl -s --fail --connect-timeout 3 http://httpd1/public")

		helpers.Sleep(2)
		cancel()
		Expect(res.CountLines()).Should(BeNumerically(">=", 3))
		filter := fmt.Sprintf("FROM %s DROP:", endpoints["httpd1"])
		Expect(res.Output().String()).Should(ContainSubstring(filter))
	})

	It("multiple monitor", func() {

		res := cilium.Exec(
			"config Debug=true DropNotification=true TraceNotification=true PolicyEnforcement=default")
		Expect(res.WasSuccessful()).Should(BeTrue())

		var monitorRes []*helpers.CmdRes
		var cancel []context.CancelFunc

		for i := 1; i <= 3; i++ {
			ctx, cancelfn := context.WithCancel(context.Background())
			monitorRes = append(monitorRes, docker.Node.ExecContext(ctx, "sudo cilium monitor"))
			cancel = append(cancel, cancelfn)
		}
		docker.SampleContainersActions("create", networkName)
		docker.ContainerExec("app1", "ping -c 5 httpd1")
		helpers.Sleep(5)
		for _, cancelfn := range cancel {
			cancelfn()
		}
		Expect(monitorRes[0].Output().String()).Should(Equal(monitorRes[1].Output().String()))
		Expect(monitorRes[0].Output().String()).Should(Equal(monitorRes[2].Output().String()))
		// Expect(monitorRes[0].Output()).Should(Equal(monitorRes[2].Output()))

		Expect(monitorRes[0].CountLines()).Should(Equal(monitorRes[1].CountLines()))
		Expect(monitorRes[0].CountLines()).Should(Equal(monitorRes[2].CountLines()))
	})
})
