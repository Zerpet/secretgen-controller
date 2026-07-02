// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	sgclient "carvel.dev/secretgen-controller/pkg/client/clientset/versioned"
	sg2client "carvel.dev/secretgen-controller/pkg/client2/clientset/versioned"
	"k8s.io/client-go/tools/clientcmd"
)

// Clients holds both generated clientsets and the resolved namespace.
type Clients struct {
	SGClient   sgclient.Interface
	SG2Client  sg2client.Interface
	Namespace  string
}

// NewClients builds Clients from kubeconfig/context/namespace flags.
// If namespace is empty the default namespace from the kubeconfig context is used.
func NewClients(kubeconfig, context, namespace string) (*Clients, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		configOverrides.CurrentContext = context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, configOverrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	resolvedNS := namespace
	if resolvedNS == "" {
		ns, _, err := clientConfig.Namespace()
		if err != nil {
			return nil, err
		}
		resolvedNS = ns
	}

	sgc, err := sgclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	sg2c, err := sg2client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &Clients{
		SGClient:  sgc,
		SG2Client: sg2c,
		Namespace: resolvedNS,
	}, nil
}
