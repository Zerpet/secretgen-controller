// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"carvel.dev/secretgen-controller/internal/cli/output"
	sgv1alpha1 "carvel.dev/secretgen-controller/pkg/apis/secretgen/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newSSHKeyCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ssh-key",
		Short:   "Manage SSHKey resources",
		Aliases: []string{"sshkey", "ssh-keys"},
	}
	cmd.AddCommand(
		newSSHKeyCreateCmd(opts),
		newSSHKeyUpdateCmd(opts),
		newSSHKeyDeleteCmd(opts),
		newSSHKeyGetCmd(opts),
		newSSHKeyDescribeCmd(opts),
		newSSHKeyListCmd(opts),
	)
	return cmd
}

func newSSHKeyCreateCmd(opts *Options) *cobra.Command {
	var secretTemplateFile string
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an SSHKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			spec := sgv1alpha1.SSHKeySpec{}
			if secretTemplateFile != "" {
				tmpl, err := loadSecretTemplateFile(secretTemplateFile)
				if err != nil {
					return err
				}
				spec.SecretTemplate = tmpl
			}
			sk := &sgv1alpha1.SSHKey{
				TypeMeta:   metav1.TypeMeta{APIVersion: "secretgen.k14s.io/v1alpha1", Kind: "SSHKey"},
				ObjectMeta: metav1.ObjectMeta{Name: args[0], Namespace: opts.ResolvedNamespace()},
				Spec:       spec,
			}
			created, err := opts.SGClient().SecretgenV1alpha1().SSHKeys(opts.ResolvedNamespace()).Create(
				context.Background(), sk, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SSHKey %q created in namespace %q\n", created.Name, created.Namespace)
			return nil
		},
	}
	cmd.Flags().StringVar(&secretTemplateFile, "secret-template-file", "", "Path to YAML file defining a secretTemplate override")
	return cmd
}

func newSSHKeyUpdateCmd(opts *Options) *cobra.Command {
	var secretTemplateFile string
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update an SSHKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			c := opts.SGClient().SecretgenV1alpha1().SSHKeys(opts.ResolvedNamespace())
			existing, err := c.Get(context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			if secretTemplateFile != "" {
				tmpl, err := loadSecretTemplateFile(secretTemplateFile)
				if err != nil {
					return err
				}
				existing.Spec.SecretTemplate = tmpl
			}
			updated, err := c.Update(context.Background(), existing, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SSHKey %q updated\n", updated.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&secretTemplateFile, "secret-template-file", "", "Path to YAML file defining a secretTemplate override")
	return cmd
}

func newSSHKeyDeleteCmd(opts *Options) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an SSHKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			name := args[0]
			if !yes {
				fmt.Printf("Delete SSHKey %q in namespace %q? [y/N] ", name, opts.ResolvedNamespace())
				reader := bufio.NewReader(os.Stdin)
				ans, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					fmt.Println("Cancelled.")
					return nil
				}
			}
			err := opts.SGClient().SecretgenV1alpha1().SSHKeys(opts.ResolvedNamespace()).Delete(
				context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SSHKey %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newSSHKeyGetCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get an SSHKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			sk, err := opts.SGClient().SecretgenV1alpha1().SSHKeys(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			p := opts.Printer()
			if p.IsTable() {
				return p.PrintTable(
					[]string{"NAME", "DESCRIPTION", "AGE"},
					[][]string{sshKeyTableRow(sk)},
				)
			}
			return p.PrintObject(sk)
		},
	}
}

func newSSHKeyDescribeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe an SSHKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			sk, err := opts.SGClient().SecretgenV1alpha1().SSHKeys(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			return describeSSHKey(opts.Printer(), sk)
		},
	}
}

func newSSHKeyListCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List SSHKey resources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			list, err := opts.SGClient().SecretgenV1alpha1().SSHKeys(opts.ResolvedNamespace()).List(
				context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			p := opts.Printer()
			if !p.IsTable() {
				return p.PrintObject(list)
			}
			rows := make([][]string, len(list.Items))
			for i := range list.Items {
				rows[i] = sshKeyTableRow(&list.Items[i])
			}
			return p.PrintTable([]string{"NAME", "DESCRIPTION", "AGE"}, rows)
		},
	}
}

func sshKeyTableRow(sk *sgv1alpha1.SSHKey) []string {
	return []string{
		sk.Name,
		sk.Status.FriendlyDescription,
		output.FormatAge(sk.CreationTimestamp),
	}
}

func describeSSHKey(p *output.Printer, sk *sgv1alpha1.SSHKey) error {
	sections := []output.DescribeSection{
		{
			Fields: []output.DescribeField{
				{Key: "Name", Value: sk.Name},
				{Key: "Namespace", Value: sk.Namespace},
				{Key: "Created", Value: output.FormatAge(sk.CreationTimestamp)},
			},
		},
		{
			Title: "Status",
			Fields: []output.DescribeField{
				{Key: "Description", Value: orNone(sk.Status.FriendlyDescription)},
				{Key: "Observed Generation", Value: fmt.Sprintf("%d", sk.Status.ObservedGeneration)},
			},
		},
	}

	if len(sk.Status.Conditions) > 0 {
		sections = append(sections, output.DescribeSection{
			Title:  "Conditions",
			Fields: conditionFields(sk.Status.Conditions),
		})
	}

	return p.PrintDescribe(sections)
}
