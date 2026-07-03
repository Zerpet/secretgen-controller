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

func newPasswordCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "password",
		Short:   "Manage Password resources",
		Aliases: []string{"pwd", "passwords"},
	}
	cmd.AddCommand(
		newPasswordCreateCmd(opts),
		newPasswordUpdateCmd(opts),
		newPasswordDeleteCmd(opts),
		newPasswordGetCmd(opts),
		newPasswordDescribeCmd(opts),
		newPasswordListCmd(opts),
	)
	return cmd
}

type passwordFlags struct {
	length           int
	digits           int
	symbols          int
	uppercase        int
	lowercase        int
	symbolCharSet    string
	secretTemplateFile string
}

func addPasswordFlags(cmd *cobra.Command, f *passwordFlags) {
	cmd.Flags().IntVar(&f.length, "length", 0, "Total password length (controller default: 40)")
	cmd.Flags().IntVar(&f.digits, "digits", 0, "Number of digit characters")
	cmd.Flags().IntVar(&f.symbols, "symbols", 0, "Number of symbol characters")
	cmd.Flags().IntVar(&f.uppercase, "uppercase", 0, "Number of uppercase letters")
	cmd.Flags().IntVar(&f.lowercase, "lowercase", 0, "Number of lowercase letters")
	cmd.Flags().StringVar(&f.symbolCharSet, "symbol-charset", "", "Set of symbols to use (controller default: !@#$%&*;.:)")
	cmd.Flags().StringVar(&f.secretTemplateFile, "secret-template-file", "", "Path to YAML file defining a secretTemplate override")
}

func applyPasswordFlags(spec *sgv1alpha1.PasswordSpec, f *passwordFlags, cmd *cobra.Command) error {
	if cmd.Flags().Changed("length") {
		spec.Length = f.length
	}
	if cmd.Flags().Changed("digits") {
		spec.Digits = f.digits
	}
	if cmd.Flags().Changed("symbols") {
		spec.Symbols = f.symbols
	}
	if cmd.Flags().Changed("uppercase") {
		spec.UppercaseLetters = f.uppercase
	}
	if cmd.Flags().Changed("lowercase") {
		spec.LowercaseLetters = f.lowercase
	}
	if cmd.Flags().Changed("symbol-charset") {
		spec.SymbolCharSet = f.symbolCharSet
	}
	if f.secretTemplateFile != "" {
		tmpl, err := loadSecretTemplateFile(f.secretTemplateFile)
		if err != nil {
			return err
		}
		spec.SecretTemplate = tmpl
	}
	return nil
}

func newPasswordCreateCmd(opts *Options) *cobra.Command {
	f := &passwordFlags{}
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Password resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			spec := sgv1alpha1.PasswordSpec{}
			if err := applyPasswordFlags(&spec, f, cmd); err != nil {
				return err
			}
			pw := &sgv1alpha1.Password{
				TypeMeta:   metav1.TypeMeta{APIVersion: "secretgen.k14s.io/v1alpha1", Kind: "Password"},
				ObjectMeta: metav1.ObjectMeta{Name: args[0], Namespace: opts.ResolvedNamespace()},
				Spec:       spec,
			}
			created, err := opts.SGClient().SecretgenV1alpha1().Passwords(opts.ResolvedNamespace()).Create(
				context.Background(), pw, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Password %q created in namespace %q\n", created.Name, created.Namespace)
			return nil
		},
	}
	addPasswordFlags(cmd, f)
	return cmd
}

func newPasswordUpdateCmd(opts *Options) *cobra.Command {
	f := &passwordFlags{}
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a Password resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			c := opts.SGClient().SecretgenV1alpha1().Passwords(opts.ResolvedNamespace())
			existing, err := c.Get(context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			if err := applyPasswordFlags(&existing.Spec, f, cmd); err != nil {
				return err
			}
			updated, err := c.Update(context.Background(), existing, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Password %q updated\n", updated.Name)
			return nil
		},
	}
	addPasswordFlags(cmd, f)
	return cmd
}

func newPasswordDeleteCmd(opts *Options) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a Password resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			name := args[0]
			if !yes {
				fmt.Printf("Delete Password %q in namespace %q? [y/N] ", name, opts.ResolvedNamespace())
				reader := bufio.NewReader(os.Stdin)
				ans, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					fmt.Println("Cancelled.")
					return nil
				}
			}
			err := opts.SGClient().SecretgenV1alpha1().Passwords(opts.ResolvedNamespace()).Delete(
				context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Password %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newPasswordGetCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get a Password resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			pw, err := opts.SGClient().SecretgenV1alpha1().Passwords(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			p := opts.Printer()
			if p.IsTable() {
				return p.PrintTable(
					[]string{"NAME", "LENGTH", "DESCRIPTION", "AGE"},
					[][]string{passwordTableRow(pw)},
				)
			}
			return p.PrintObject(pw)
		},
	}
}

func newPasswordDescribeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Password resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			pw, err := opts.SGClient().SecretgenV1alpha1().Passwords(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			return describePassword(opts.Printer(), pw)
		},
	}
}

func newPasswordListCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Password resources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			list, err := opts.SGClient().SecretgenV1alpha1().Passwords(opts.ResolvedNamespace()).List(
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
				rows[i] = passwordTableRow(&list.Items[i])
			}
			return p.PrintTable([]string{"NAME", "LENGTH", "DESCRIPTION", "AGE"}, rows)
		},
	}
}

func passwordTableRow(pw *sgv1alpha1.Password) []string {
	return []string{
		pw.Name,
		fmt.Sprintf("%d", pw.Spec.Length),
		pw.Status.FriendlyDescription,
		output.FormatAge(pw.CreationTimestamp),
	}
}

func describePassword(p *output.Printer, pw *sgv1alpha1.Password) error {
	symbolCharSet := pw.Spec.SymbolCharSet
	if symbolCharSet == "" {
		symbolCharSet = "!@#$%&*;.: (default)"
	}

	sections := []output.DescribeSection{
		{
			Fields: []output.DescribeField{
				{Key: "Name", Value: pw.Name},
				{Key: "Namespace", Value: pw.Namespace},
				{Key: "Created", Value: output.FormatAge(pw.CreationTimestamp)},
			},
		},
		{
			Title: "Spec",
			Fields: []output.DescribeField{
				{Key: "Length", Value: fmt.Sprintf("%d", pw.Spec.Length)},
				{Key: "Digits", Value: fmt.Sprintf("%d", pw.Spec.Digits)},
				{Key: "Symbols", Value: fmt.Sprintf("%d", pw.Spec.Symbols)},
				{Key: "Uppercase Letters", Value: fmt.Sprintf("%d", pw.Spec.UppercaseLetters)},
				{Key: "Lowercase Letters", Value: fmt.Sprintf("%d", pw.Spec.LowercaseLetters)},
				{Key: "Symbol Char Set", Value: symbolCharSet},
			},
		},
		{
			Title: "Status",
			Fields: []output.DescribeField{
				{Key: "Description", Value: orNone(pw.Status.FriendlyDescription)},
				{Key: "Observed Generation", Value: fmt.Sprintf("%d", pw.Status.ObservedGeneration)},
			},
		},
	}

	if len(pw.Status.Conditions) > 0 {
		sections = append(sections, output.DescribeSection{
			Title:  "Conditions",
			Fields: conditionFields(pw.Status.Conditions),
		})
	}

	return p.PrintDescribe(sections)
}
