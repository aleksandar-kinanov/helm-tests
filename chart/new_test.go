package helm_tester

import (
	"path/filepath"
	"strings"
	"testing"

	// appsv1 "k8s.io/api/apps/v1"
	// batchv1 "k8s.io/api/batch/v1"
	// "k8s.io/api/batch/v1beta1"
	// corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	// rbacv1 "k8s.io/api/rbac/v1"

	"github.com/gruntwork-io/terratest/modules/helm"
	// "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	// monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
)

var (
	err           error
	helmChartPath string
	releaseName   = "helm-basic"
	namespaceName = "crdb-" + strings.ToLower(random.UniqueId())
)

func init() {
	helmChartPath, err = filepath.Abs("./nginx-example")
	if err != nil {
		panic(err)
	}
}

func TestBasicNginxConfig(t *testing.T) {
	t.Parallel()
	var ingress networkingv1.Ingress
	testCases := []struct {
		name   string
		values map[string]string
		expect networkingv1.IngressSpec
	}{
		// {
		// 	"Self Signer and cert manager set to false",
		// 	map[string]string{
		// 		"tls.certs.selfSigner.enabled": "false",
		// 		"tls.certs.certManager":        "false",
		// 	},
		// 	"You have to enable either self signed certificates or certificate manager, if you have enabled tls",
		// },
		{
			"Self Signer and cert manager set to true",
			map[string]string{
				"ingress.hosts[0].host":     "dummy-host",
				"ingress.hosts[0].paths[0]": "ASDADASDASDA",
				"tls.certs.certManager":     "true",
			},
			networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: "dummy-host",
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path: "ASDADASDASDA",
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "helm-basic-nginx-example",
												Port: networkingv1.ServiceBackendPort{
													Number: 80,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		// Here, we capture the range variable and force it into the scope of this block. If we don't do this, when the
		// subtest switches contexts (because of t.Parallel), the testCase value will have been updated by the for loop
		// and will be the next testCase!
		testCase := testCase
		t.Run(testCase.name, func(subT *testing.T) {
			subT.Parallel()

			// Now we try rendering the template, but verify we get an error
			options := &helm.Options{SetValues: testCase.values}
			output, err := helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/ingress.yaml"})
			if err != nil {
				t.Logf("%v\n", err)
			}
			helm.UnmarshalK8SYaml(t, output, &ingress)
			t.Logf("%v\n", ingress.Spec)
			require.Equal(subT, testCase.expect, ingress.Spec)

		})
	}

}
