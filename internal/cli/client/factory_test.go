// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"testing"

	"carvel.dev/secretgen-controller/internal/cli/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// minimalKubeconfig returns a minimal kubeconfig pointing at a fake server.
func minimalKubeconfig(t *testing.T, namespace string) string {
	t.Helper()
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters["test-cluster"] = &clientcmdapi.Cluster{
		Server: "https://fake-server:6443",
	}
	cfg.AuthInfos["test-user"] = &clientcmdapi.AuthInfo{}
	cfg.Contexts["test-context"] = &clientcmdapi.Context{
		Cluster:   "test-cluster",
		AuthInfo:  "test-user",
		Namespace: namespace,
	}
	cfg.CurrentContext = "test-context"

	tmpFile := t.TempDir() + "/kubeconfig"
	err := clientcmd.WriteToFile(*cfg, tmpFile)
	require.NoError(t, err)
	return tmpFile
}

func TestNewClients_ResolvedNamespaceFromFlag(t *testing.T) {
	kubeconfigPath := minimalKubeconfig(t, "context-ns")

	c, err := client.NewClients(kubeconfigPath, "", "override-ns")
	require.NoError(t, err)

	assert.Equal(t, "override-ns", c.Namespace)
	assert.NotNil(t, c.SGClient)
	assert.NotNil(t, c.SG2Client)
}

func TestNewClients_ResolvedNamespaceFromContext(t *testing.T) {
	kubeconfigPath := minimalKubeconfig(t, "context-ns")

	c, err := client.NewClients(kubeconfigPath, "", "")
	require.NoError(t, err)

	assert.Equal(t, "context-ns", c.Namespace)
}

func TestNewClients_CustomContext(t *testing.T) {
	t.Helper()
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters["cluster1"] = &clientcmdapi.Cluster{Server: "https://fake1:6443"}
	cfg.Clusters["cluster2"] = &clientcmdapi.Cluster{Server: "https://fake2:6443"}
	cfg.AuthInfos["user"] = &clientcmdapi.AuthInfo{}
	cfg.Contexts["ctx1"] = &clientcmdapi.Context{Cluster: "cluster1", AuthInfo: "user", Namespace: "ns1"}
	cfg.Contexts["ctx2"] = &clientcmdapi.Context{Cluster: "cluster2", AuthInfo: "user", Namespace: "ns2"}
	cfg.CurrentContext = "ctx1"

	tmpFile := t.TempDir() + "/kubeconfig"
	err := clientcmd.WriteToFile(*cfg, tmpFile)
	require.NoError(t, err)

	c, err := client.NewClients(tmpFile, "ctx2", "")
	require.NoError(t, err)

	assert.Equal(t, "ns2", c.Namespace)
}

func TestNewClients_MissingKubeconfig(t *testing.T) {
	_, err := client.NewClients("/nonexistent/kubeconfig", "", "")
	assert.Error(t, err)
}
