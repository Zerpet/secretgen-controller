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

func newRSAKeyCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rsa-key",
		Short:   "Manage RSAKey resources",
		Aliases: []string{"rsakey", "rsa-keys"},
	}
	cmd.AddCommand(
		newRSAKeyCreateCmd(opts),
		newRSAKeyUpdateCmd(opts),
		newRSAKeyDeleteCmd(opts),
		newRSAKeyGetCmd(opts),
		newRSAKeyDescribeCmd(opts),
		newRSAKeyListCmd(opts),
	)
	return cmd
}

func newRSAKeyCreateCmd(opts *Options) *cobra.Command {
	var secretTemplateFile string
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an RSAKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			spec := sgv1alpha1.RSAKeySpec{}
			if secretTemplateFile != "" {
				tmpl, err := loadSecretTemplateFile(secretTemplateFile)
				if err != nil {
					return err
				}
				spec.SecretTemplate = tmpl
			}
			rk := &sgv1alpha1.RSAKey{
				TypeMeta:   metav1.TypeMeta{APIVersion: "secretgen.k14s.io/v1alpha1", Kind: "RSAKey"},
				ObjectMeta: metav1.ObjectMeta{Name: args[0], Namespace: opts.ResolvedNamespace()},
				Spec:       spec,
			}
			created, err := opts.SGClient().SecretgenV1alpha1().RSAKeys(opts.ResolvedNamespace()).Create(
				context.Background(), rk, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("RSAKey %q created in namespace %q\n", created.Name, created.Namespace)
			return nil
		},
	}
	cmd.Flags().StringVar(&secretTemplateFile, "secret-template-file", "", "Path to YAML file defining a secretTemplate override")
	return cmd
}

func newRSAKeyUpdateCmd(opts *Options) *cobra.Command {
	var secretTemplateFile string
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update an RSAKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			c := opts.SGClient().SecretgenV1alpha1().RSAKeys(opts.ResolvedNamespace())
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
			fmt.Printf("RSAKey %q updated\n", updated.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&secretTemplateFile, "secret-template-file", "", "Path to YAML file defining a secretTemplate override")
	return cmd
}

func newRSAKeyDeleteCmd(opts *Options) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an RSAKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			name := args[0]
			if !yes {
				fmt.Printf("Delete RSAKey %q in namespace %q? [y/N] ", name, opts.ResolvedNamespace())
				reader := bufio.NewReader(os.Stdin)
				ans, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					fmt.Println("Cancelled.")
					return nil
				}
			}
			err := opts.SGClient().SecretgenV1alpha1().RSAKeys(opts.ResolvedNamespace()).Delete(
				context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("RSAKey %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newRSAKeyGetCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get an RSAKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			rk, err := opts.SGClient().SecretgenV1alpha1().RSAKeys(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			p := opts.Printer()
			if p.IsTable() {
				return p.PrintTable(
					[]string{"NAME", "DESCRIPTION", "AGE"},
					[][]string{rsaKeyTableRow(rk)},
				)
			}
			return p.PrintObject(rk)
		},
	}
}

func newRSAKeyDescribeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe an RSAKey resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			rk, err := opts.SGClient().SecretgenV1alpha1().RSAKeys(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			return describeRSAKey(opts.Printer(), rk)
		},
	}
}

func newRSAKeyListCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List RSAKey resources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			list, err := opts.SGClient().SecretgenV1alpha1().RSAKeys(opts.ResolvedNamespace()).List(
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
				rows[i] = rsaKeyTableRow(&list.Items[i])
			}
			return p.PrintTable([]string{"NAME", "DESCRIPTION", "AGE"}, rows)
		},
	}
}

func rsaKeyTableRow(rk *sgv1alpha1.RSAKey) []string {
	return []string{
		rk.Name,
		rk.Status.FriendlyDescription,
		output.FormatAge(rk.CreationTimestamp),
	}
}

func describeRSAKey(p *output.Printer, rk *sgv1alpha1.RSAKey) error {
	sections := []output.DescribeSection{
		{
			Fields: []output.DescribeField{
				{Key: "Name", Value: rk.Name},
				{Key: "Namespace", Value: rk.Namespace},
				{Key: "Created", Value: output.FormatAge(rk.CreationTimestamp)},
			},
		},
		{
			Title: "Status",
			Fields: []output.DescribeField{
				{Key: "Description", Value: orNone(rk.Status.FriendlyDescription)},
				{Key: "Observed Generation", Value: fmt.Sprintf("%d", rk.Status.ObservedGeneration)},
			},
		},
	}

	if len(rk.Status.Conditions) > 0 {
		sections = append(sections, output.DescribeSection{
			Title:  "Conditions",
			Fields: conditionFields(rk.Status.Conditions),
		})
	}

	return p.PrintDescribe(sections)
}
