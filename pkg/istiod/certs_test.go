// Copyright 2019 Istio Authors
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

package istiod

import (
	"testing"
	"time"
)

// Generate a certificate. Typical time ~500ms
func TestCerts(t *testing.T) {

	t0 := time.Now()
	client, kcfg, err := CreateClientset("", "")
	if err != nil || kcfg.TLSClientConfig.Insecure {
		t.Skip("Missing K8S", err)
	}

	certChain, keyPEM, err := GenKeyCertK8sCA(client.CertificatesV1beta1(), "istio-system", "istio-pilot.istio-system")
	if err != nil {
		t.Fatal("Fail to generate cert", err)
	}

	caCert := kcfg.TLSClientConfig.CAData

	t.Log("Cert Chain:", time.Since(t0), "\n", string(certChain))
	t.Log("Key\n", string(keyPEM))

	err = CheckCert(certChain, caCert)
	if err != nil {
		t.Fatal("Invalid root CA or cert", err)
	}
}

func BenchmarkCerts(t *testing.B) {
	client, kcfg, err := CreateClientset("", "")
	if err != nil || kcfg.TLSClientConfig.Insecure {
		t.Skip("Missing K8S", err)
	}
	for n := 0; n < t.N; n++ {
		certChain, _, err := GenKeyCertK8sCA(client.CertificatesV1beta1(), "istio-system", "istio-pilot.istio-system")
		if err != nil {
			t.Fatal("Fail to generate cert", err)
		}

		caCert := kcfg.TLSClientConfig.CAData
		_ = append(certChain, caCert...)
	}
}
