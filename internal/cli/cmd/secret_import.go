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
	sg2v1alpha1 "carvel.dev/secretgen-controller/pkg/apis/secretgen2/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newSecretImportCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret-import",
		Short:   "Manage SecretImport resources",
		Aliases: []string{"secimp", "secret-imports"},
	}
	cmd.AddCommand(
		newSecretImportCreateCmd(opts),
		newSecretImportUpdateCmd(opts),
		newSecretImportDeleteCmd(opts),
		newSecretImportGetCmd(opts),
		newSecretImportDescribeCmd(opts),
		newSecretImportListCmd(opts),
	)
	return cmd
}

func newSecretImportCreateCmd(opts *Options) *cobra.Command {
	var fromNamespace string
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a SecretImport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			si := &sg2v1alpha1.SecretImport{
				TypeMeta:   metav1.TypeMeta{APIVersion: "secretgen.carvel.dev/v1alpha1", Kind: "SecretImport"},
				ObjectMeta: metav1.ObjectMeta{Name: args[0], Namespace: opts.ResolvedNamespace()},
				Spec:       sg2v1alpha1.SecretImportSpec{FromNamespace: fromNamespace},
			}
			if err := si.Validate(); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			created, err := opts.SG2Client().SecretgenV1alpha1().SecretImports(opts.ResolvedNamespace()).Create(
				context.Background(), si, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretImport %q created in namespace %q\n", created.Name, created.Namespace)
			return nil
		},
	}
	cmd.Flags().StringVar(&fromNamespace, "from-namespace", "", "Source namespace to import the secret from (required)")
	_ = cmd.MarkFlagRequired("from-namespace")
	return cmd
}

func newSecretImportUpdateCmd(opts *Options) *cobra.Command {
	var fromNamespace string
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a SecretImport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			c := opts.SG2Client().SecretgenV1alpha1().SecretImports(opts.ResolvedNamespace())
			existing, err := c.Get(context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("from-namespace") {
				existing.Spec.FromNamespace = fromNamespace
			}
			if err := existing.Validate(); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			updated, err := c.Update(context.Background(), existing, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretImport %q updated\n", updated.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&fromNamespace, "from-namespace", "", "Source namespace to import the secret from")
	return cmd
}

func newSecretImportDeleteCmd(opts *Options) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a SecretImport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			name := args[0]
			if !yes {
				fmt.Printf("Delete SecretImport %q in namespace %q? [y/N] ", name, opts.ResolvedNamespace())
				reader := bufio.NewReader(os.Stdin)
				ans, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					fmt.Println("Cancelled.")
					return nil
				}
			}
			err := opts.SG2Client().SecretgenV1alpha1().SecretImports(opts.ResolvedNamespace()).Delete(
				context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretImport %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newSecretImportGetCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get a SecretImport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			si, err := opts.SG2Client().SecretgenV1alpha1().SecretImports(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			p := opts.Printer()
			if p.IsTable() {
				return p.PrintTable(
					[]string{"NAME", "FROM-NAMESPACE", "DESCRIPTION", "AGE"},
					[][]string{secretImportTableRow(si)},
				)
			}
			return p.PrintObject(si)
		},
	}
}

func newSecretImportDescribeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a SecretImport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			si, err := opts.SG2Client().SecretgenV1alpha1().SecretImports(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			return describeSecretImport(opts.Printer(), si)
		},
	}
}

func newSecretImportListCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List SecretImport resources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			list, err := opts.SG2Client().SecretgenV1alpha1().SecretImports(opts.ResolvedNamespace()).List(
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
				rows[i] = secretImportTableRow(&list.Items[i])
			}
			return p.PrintTable([]string{"NAME", "FROM-NAMESPACE", "DESCRIPTION", "AGE"}, rows)
		},
	}
}

func secretImportTableRow(si *sg2v1alpha1.SecretImport) []string {
	return []string{
		si.Name,
		si.Spec.FromNamespace,
		si.Status.FriendlyDescription,
		output.FormatAge(si.CreationTimestamp),
	}
}

func describeSecretImport(p *output.Printer, si *sg2v1alpha1.SecretImport) error {
	sections := []output.DescribeSection{
		{
			Fields: []output.DescribeField{
				{Key: "Name", Value: si.Name},
				{Key: "Namespace", Value: si.Namespace},
				{Key: "Created", Value: output.FormatAge(si.CreationTimestamp)},
			},
		},
		{
			Title: "Spec",
			Fields: []output.DescribeField{
				{Key: "From Namespace", Value: orNone(si.Spec.FromNamespace)},
			},
		},
		{
			Title: "Status",
			Fields: []output.DescribeField{
				{Key: "Description", Value: orNone(si.Status.FriendlyDescription)},
				{Key: "Observed Generation", Value: fmt.Sprintf("%d", si.Status.ObservedGeneration)},
			},
		},
	}

	if len(si.Status.Conditions) > 0 {
		sections = append(sections, output.DescribeSection{
			Title:  "Conditions",
			Fields: conditionFields(si.Status.Conditions),
		})
	}

	return p.PrintDescribe(sections)
}
