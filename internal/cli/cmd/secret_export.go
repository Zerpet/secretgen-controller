// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"carvel.dev/secretgen-controller/internal/cli/output"
	sg2v1alpha1 "carvel.dev/secretgen-controller/pkg/apis/secretgen2/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newSecretExportCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret-export",
		Short:   "Manage SecretExport resources",
		Aliases: []string{"secexp", "secret-exports"},
	}
	cmd.AddCommand(
		newSecretExportCreateCmd(opts),
		newSecretExportUpdateCmd(opts),
		newSecretExportDeleteCmd(opts),
		newSecretExportGetCmd(opts),
		newSecretExportDescribeCmd(opts),
		newSecretExportListCmd(opts),
	)
	return cmd
}

type secretExportFlags struct {
	toNamespace  string
	toNamespaces []string
	toSelector   string // raw JSON array for dangerousToNamespacesSelector
}

func addSecretExportFlags(cmd *cobra.Command, f *secretExportFlags) {
	cmd.Flags().StringVar(&f.toNamespace, "to-namespace", "", "Single destination namespace (use * for all)")
	cmd.Flags().StringSliceVar(&f.toNamespaces, "to-namespaces", nil, "Comma-separated list of destination namespaces")
	cmd.Flags().StringVar(&f.toSelector, "to-selector", "",
		`JSON array of namespace selector match fields, e.g. '[{"key":"metadata.labels.env","operator":"In","values":["prod"]}]'`)
}

func applySecretExportFlags(spec *sg2v1alpha1.SecretExportSpec, f *secretExportFlags, cmd *cobra.Command) error {
	if cmd.Flags().Changed("to-namespace") {
		spec.ToNamespace = f.toNamespace
	}
	if cmd.Flags().Changed("to-namespaces") {
		spec.ToNamespaces = f.toNamespaces
	}
	if cmd.Flags().Changed("to-selector") {
		var selectors []sg2v1alpha1.SelectorMatchField
		if err := json.Unmarshal([]byte(f.toSelector), &selectors); err != nil {
			return fmt.Errorf("parsing --to-selector: %w", err)
		}
		spec.ToNamespacesSelector = selectors
	}
	return nil
}

func newSecretExportCreateCmd(opts *Options) *cobra.Command {
	f := &secretExportFlags{}
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a SecretExport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			spec := sg2v1alpha1.SecretExportSpec{}
			if err := applySecretExportFlags(&spec, f, cmd); err != nil {
				return err
			}
			se := &sg2v1alpha1.SecretExport{
				TypeMeta:   metav1.TypeMeta{APIVersion: "secretgen.carvel.dev/v1alpha1", Kind: "SecretExport"},
				ObjectMeta: metav1.ObjectMeta{Name: args[0], Namespace: opts.ResolvedNamespace()},
				Spec:       spec,
			}
			if err := se.Validate(); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			created, err := opts.SG2Client().SecretgenV1alpha1().SecretExports(opts.ResolvedNamespace()).Create(
				context.Background(), se, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretExport %q created in namespace %q\n", created.Name, created.Namespace)
			return nil
		},
	}
	addSecretExportFlags(cmd, f)
	return cmd
}

func newSecretExportUpdateCmd(opts *Options) *cobra.Command {
	f := &secretExportFlags{}
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a SecretExport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			c := opts.SG2Client().SecretgenV1alpha1().SecretExports(opts.ResolvedNamespace())
			existing, err := c.Get(context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			if err := applySecretExportFlags(&existing.Spec, f, cmd); err != nil {
				return err
			}
			if err := existing.Validate(); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			updated, err := c.Update(context.Background(), existing, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretExport %q updated\n", updated.Name)
			return nil
		},
	}
	addSecretExportFlags(cmd, f)
	return cmd
}

func newSecretExportDeleteCmd(opts *Options) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a SecretExport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			name := args[0]
			if !yes {
				fmt.Printf("Delete SecretExport %q in namespace %q? [y/N] ", name, opts.ResolvedNamespace())
				reader := bufio.NewReader(os.Stdin)
				ans, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					fmt.Println("Cancelled.")
					return nil
				}
			}
			err := opts.SG2Client().SecretgenV1alpha1().SecretExports(opts.ResolvedNamespace()).Delete(
				context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretExport %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newSecretExportGetCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get a SecretExport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			se, err := opts.SG2Client().SecretgenV1alpha1().SecretExports(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			p := opts.Printer()
			if p.IsTable() {
				return p.PrintTable(
					[]string{"NAME", "TO-NAMESPACES", "DESCRIPTION", "AGE"},
					[][]string{secretExportTableRow(se)},
				)
			}
			return p.PrintObject(se)
		},
	}
}

func newSecretExportDescribeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a SecretExport resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			se, err := opts.SG2Client().SecretgenV1alpha1().SecretExports(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			return describeSecretExport(opts.Printer(), se)
		},
	}
}

func newSecretExportListCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List SecretExport resources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			list, err := opts.SG2Client().SecretgenV1alpha1().SecretExports(opts.ResolvedNamespace()).List(
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
				rows[i] = secretExportTableRow(&list.Items[i])
			}
			return p.PrintTable([]string{"NAME", "TO-NAMESPACES", "DESCRIPTION", "AGE"}, rows)
		},
	}
}

func secretExportTableRow(se *sg2v1alpha1.SecretExport) []string {
	toNSes := se.StaticToNamespaces()
	return []string{
		se.Name,
		output.JoinStrings(toNSes, ", "),
		se.Status.FriendlyDescription,
		output.FormatAge(se.CreationTimestamp),
	}
}

func describeSecretExport(p *output.Printer, se *sg2v1alpha1.SecretExport) error {
	toNSes := se.StaticToNamespaces()

	selectorStr := "<none>"
	if len(se.Spec.ToNamespacesSelector) > 0 {
		b, _ := json.Marshal(se.Spec.ToNamespacesSelector)
		selectorStr = string(b)
	}

	sections := []output.DescribeSection{
		{
			Fields: []output.DescribeField{
				{Key: "Name", Value: se.Name},
				{Key: "Namespace", Value: se.Namespace},
				{Key: "Created", Value: output.FormatAge(se.CreationTimestamp)},
			},
		},
		{
			Title: "Spec",
			Fields: []output.DescribeField{
				{Key: "To Namespaces", Value: output.JoinStrings(toNSes, ", ")},
				{Key: "To Namespaces Selector", Value: selectorStr},
			},
		},
		{
			Title: "Status",
			Fields: []output.DescribeField{
				{Key: "Description", Value: orNone(se.Status.FriendlyDescription)},
				{Key: "Observed Generation", Value: fmt.Sprintf("%d", se.Status.ObservedGeneration)},
			},
		},
	}

	if len(se.Status.Conditions) > 0 {
		sections = append(sections, output.DescribeSection{
			Title:  "Conditions",
			Fields: conditionFields(se.Status.Conditions),
		})
	}

	return p.PrintDescribe(sections)
}
