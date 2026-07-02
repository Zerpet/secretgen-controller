// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"carvel.dev/secretgen-controller/internal/cli/client"
	"carvel.dev/secretgen-controller/internal/cli/output"
	sgclient "carvel.dev/secretgen-controller/pkg/client/clientset/versioned"
	sg2client "carvel.dev/secretgen-controller/pkg/client2/clientset/versioned"
	"github.com/spf13/cobra"
)

// Options holds global CLI flags and lazily-initialised cluster clients.
type Options struct {
	Kubeconfig string
	KubeCtx    string
	Namespace  string
	Output     string

	clients *client.Clients
}

// Connect builds the cluster clients on first call (idempotent).
func (o *Options) Connect() error {
	if o.clients != nil {
		return nil
	}
	c, err := client.NewClients(o.Kubeconfig, o.KubeCtx, o.Namespace)
	if err != nil {
		return fmt.Errorf("connecting to cluster: %w", err)
	}
	o.clients = c
	return nil
}

// SGClient returns the secretgen/v1alpha1 client (call Connect first).
func (o *Options) SGClient() sgclient.Interface { return o.clients.SGClient }

// SG2Client returns the secretgen.carvel.dev/v1alpha1 client (call Connect first).
func (o *Options) SG2Client() sg2client.Interface { return o.clients.SG2Client }

// ResolvedNamespace returns the namespace resolved by Connect.
func (o *Options) ResolvedNamespace() string { return o.clients.Namespace }

// Printer returns a Printer configured by the --output flag.
func (o *Options) Printer() *output.Printer {
	return output.NewPrinter(o.Output, os.Stdout)
}

// Execute is the CLI entrypoint called by cmd/cli/main.go.
func Execute(version string) {
	opts := &Options{}

	root := &cobra.Command{
		Use:   "sgctl",
		Short: "CLI for managing secretgen-controller resources",
		Long: `sgctl manages Kubernetes resources provided by secretgen-controller:
Certificates, Passwords, RSA Keys, SSH Keys, SecretExports, SecretImports, and SecretTemplates.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&opts.Kubeconfig, "kubeconfig", "", "Path to kubeconfig file (defaults to KUBECONFIG env or ~/.kube/config)")
	root.PersistentFlags().StringVar(&opts.KubeCtx, "context", "", "Kubernetes context to use")
	root.PersistentFlags().StringVarP(&opts.Namespace, "namespace", "n", "", "Kubernetes namespace (defaults to current context namespace)")
	root.PersistentFlags().StringVarP(&opts.Output, "output", "o", "table", "Output format: table, json, yaml")

	root.AddCommand(newVersionCmd(version))
	root.AddCommand(newCertificateCmd(opts))
	root.AddCommand(newPasswordCmd(opts))
	root.AddCommand(newRSAKeyCmd(opts))
	root.AddCommand(newSSHKeyCmd(opts))
	root.AddCommand(newSecretExportCmd(opts))
	root.AddCommand(newSecretImportCmd(opts))
	root.AddCommand(newSecretTemplateCmd(opts))

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
