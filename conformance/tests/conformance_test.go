//go:build conformance

/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package tests

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/yaml"
)

func TestConformance(t *testing.T) {
	g := NewGomegaWithT(t)
	cfg, err := config.GetConfig()
	g.Expect(err).To(BeNil())

	client, err := client.New(cfg, client.Options{})
	g.Expect(err).To(BeNil())

	g.Expect(v1alpha2.AddToScheme(client.Scheme())).To(Succeed())
	g.Expect(v1beta1.AddToScheme(client.Scheme())).To(Succeed())

	supportedFeatures := suite.ParseSupportedFeatures(*flags.SupportedFeatures)
	exemptFeatures := parseSupportedFeatures(*flags.ExemptFeatures)

	t.Logf(`Running conformance tests with %s GatewayClass\n cleanup: %t\n`+
		`debug: %t\n enable all features: %t \n supported features: [%v]\n exempt features: [%v]`,
		*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug,
		*flags.EnableAllSupportedFeatures, *flags.SupportedFeatures, *flags.ExemptFeatures)

	expSuite, err := suite.NewExperimentalConformanceTestSuite(suite.ExperimentalConformanceOptions{
		Options: suite.Options{
			Client:                     client,
			GatewayClassName:           *flags.GatewayClassName,
			Debug:                      *flags.ShowDebug,
			CleanupBaseResources:       *flags.CleanupBaseResources,
			SupportedFeatures:          supportedFeatures,
			ExemptFeatures:             exemptFeatures,
			EnableAllSupportedFeatures: *flags.EnableAllSupportedFeatures,
		},
		Implementation: v1alpha1.Implementation{
			Organization: "nginxinc",
			Project:      "nginx-kubernetes-gateway",
			URL:          "https://github.com/nginxinc/nginx-kubernetes-gateway",
			Version:      "v0.5.0",
			Contact: []string{
				"@nginxinc/kubernetes-gateway",
			},
		},
		ConformanceProfiles: sets.New(suite.HTTPConformanceProfileName),
	})
	g.Expect(err).To(Not(HaveOccurred()))

	expSuite.Setup(t)

	err = expSuite.Run(t, tests.ConformanceTests)
	g.Expect(err).To(Not(HaveOccurred()))

	report, err := expSuite.Report()
	g.Expect(err).To(Not(HaveOccurred()))

	yamlReport, err := yaml.Marshal(report)
	g.Expect(err).ToNot(HaveOccurred())

	t.Logf("Conformance Profile: \n%s\n", string(yamlReport))
}

// parseSupportedFeatures parses flag arguments and converts the string to
// sets.Set[suite.SupportedFeature]
// FIXME(kate-osborn): Use exported ParseSupportedFeatures function
// https://github.com/kubernetes-sigs/gateway-api/blob/63e423cf1b837991d2747742199d90863a98b0c3/conformance/utils/suite/suite.go#L235
// once it's released. https://github.com/nginxinc/nginx-kubernetes-gateway/issues/779
func parseSupportedFeatures(f string) sets.Set[suite.SupportedFeature] {
	if f == "" {
		return nil
	}
	res := sets.Set[suite.SupportedFeature]{}
	for _, value := range strings.Split(f, ",") {
		res.Insert(suite.SupportedFeature(value))
	}
	return res
}
