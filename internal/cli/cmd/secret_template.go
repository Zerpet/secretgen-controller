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
	sigsyaml "sigs.k8s.io/yaml"
)

func newSecretTemplateCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret-template",
		Short:   "Manage SecretTemplate resources",
		Aliases: []string{"tmpl", "secret-templates"},
	}
	cmd.AddCommand(
		newSecretTemplateCreateCmd(opts),
		newSecretTemplateUpdateCmd(opts),
		newSecretTemplateDeleteCmd(opts),
		newSecretTemplateGetCmd(opts),
		newSecretTemplateDescribeCmd(opts),
		newSecretTemplateListCmd(opts),
	)
	return cmd
}

// loadSecretTemplateSpec reads a YAML file that is either a full SecretTemplate CR
// or just a SecretTemplateSpec, and returns the spec.
func loadSecretTemplateSpec(path string) (sg2v1alpha1.SecretTemplateSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return sg2v1alpha1.SecretTemplateSpec{}, fmt.Errorf("reading file: %w", err)
	}

	// Try full CR first (look for "spec:" key at the document level).
	var full sg2v1alpha1.SecretTemplate
	if err := sigsyaml.Unmarshal(b, &full); err != nil {
		return sg2v1alpha1.SecretTemplateSpec{}, fmt.Errorf("parsing file: %w", err)
	}
	if len(full.Spec.InputResources) > 0 || full.Spec.JSONPathTemplate != nil || full.Spec.ServiceAccountName != "" {
		return full.Spec, nil
	}

	// Fall back to treating the file as a bare spec object.
	var spec sg2v1alpha1.SecretTemplateSpec
	if err := sigsyaml.Unmarshal(b, &spec); err != nil {
		return sg2v1alpha1.SecretTemplateSpec{}, fmt.Errorf("parsing file as SecretTemplateSpec: %w", err)
	}
	return spec, nil
}

func newSecretTemplateCreateCmd(opts *Options) *cobra.Command {
	var (
		file               string
		serviceAccountName string
	)
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a SecretTemplate resource",
		Long: `Create a SecretTemplate resource.

The --file flag accepts a YAML file that is either a full SecretTemplate CR or a bare SecretTemplateSpec.
The --service-account flag sets spec.serviceAccountName and overrides any value in the file.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			var spec sg2v1alpha1.SecretTemplateSpec
			if file != "" {
				var err error
				spec, err = loadSecretTemplateSpec(file)
				if err != nil {
					return err
				}
			}
			if cmd.Flags().Changed("service-account") {
				spec.ServiceAccountName = serviceAccountName
			}
			st := &sg2v1alpha1.SecretTemplate{
				TypeMeta:   metav1.TypeMeta{APIVersion: "secretgen.carvel.dev/v1alpha1", Kind: "SecretTemplate"},
				ObjectMeta: metav1.ObjectMeta{Name: args[0], Namespace: opts.ResolvedNamespace()},
				Spec:       spec,
			}
			created, err := opts.SG2Client().SecretgenV1alpha1().SecretTemplates(opts.ResolvedNamespace()).Create(
				context.Background(), st, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretTemplate %q created in namespace %q\n", created.Name, created.Namespace)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file containing a SecretTemplate spec")
	cmd.Flags().StringVar(&serviceAccountName, "service-account", "", "ServiceAccount name used to read input resources")
	return cmd
}

func newSecretTemplateUpdateCmd(opts *Options) *cobra.Command {
	var (
		file               string
		serviceAccountName string
	)
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a SecretTemplate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			c := opts.SG2Client().SecretgenV1alpha1().SecretTemplates(opts.ResolvedNamespace())
			existing, err := c.Get(context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			if file != "" {
				spec, err := loadSecretTemplateSpec(file)
				if err != nil {
					return err
				}
				existing.Spec = spec
			}
			if cmd.Flags().Changed("service-account") {
				existing.Spec.ServiceAccountName = serviceAccountName
			}
			updated, err := c.Update(context.Background(), existing, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretTemplate %q updated\n", updated.Name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file containing a SecretTemplate spec")
	cmd.Flags().StringVar(&serviceAccountName, "service-account", "", "ServiceAccount name used to read input resources")
	return cmd
}

func newSecretTemplateDeleteCmd(opts *Options) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a SecretTemplate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			name := args[0]
			if !yes {
				fmt.Printf("Delete SecretTemplate %q in namespace %q? [y/N] ", name, opts.ResolvedNamespace())
				reader := bufio.NewReader(os.Stdin)
				ans, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					fmt.Println("Cancelled.")
					return nil
				}
			}
			err := opts.SG2Client().SecretgenV1alpha1().SecretTemplates(opts.ResolvedNamespace()).Delete(
				context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("SecretTemplate %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newSecretTemplateGetCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get a SecretTemplate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			st, err := opts.SG2Client().SecretgenV1alpha1().SecretTemplates(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			p := opts.Printer()
			if p.IsTable() {
				return p.PrintTable(
					[]string{"NAME", "SECRET", "DESCRIPTION", "AGE"},
					[][]string{secretTemplateTableRow(st)},
				)
			}
			return p.PrintObject(st)
		},
	}
}

func newSecretTemplateDescribeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a SecretTemplate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			st, err := opts.SG2Client().SecretgenV1alpha1().SecretTemplates(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			return describeSecretTemplate(opts.Printer(), st)
		},
	}
}

func newSecretTemplateListCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List SecretTemplate resources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			list, err := opts.SG2Client().SecretgenV1alpha1().SecretTemplates(opts.ResolvedNamespace()).List(
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
				rows[i] = secretTemplateTableRow(&list.Items[i])
			}
			return p.PrintTable([]string{"NAME", "SECRET", "DESCRIPTION", "AGE"}, rows)
		},
	}
}

func secretTemplateTableRow(st *sg2v1alpha1.SecretTemplate) []string {
	secretName := st.Status.Secret.Name
	if secretName == "" {
		secretName = "<pending>"
	}
	return []string{
		st.Name,
		secretName,
		st.Status.FriendlyDescription,
		output.FormatAge(st.CreationTimestamp),
	}
}

func describeSecretTemplate(p *output.Printer, st *sg2v1alpha1.SecretTemplate) error {
	inputResources := "<none>"
	if len(st.Spec.InputResources) > 0 {
		names := make([]string, len(st.Spec.InputResources))
		for i, ir := range st.Spec.InputResources {
			names[i] = fmt.Sprintf("%s (%s/%s %s)", ir.Name, ir.Ref.APIVersion, ir.Ref.Kind, ir.Ref.Name)
		}
		inputResources = strings.Join(names, "\n                               ")
	}

	secretName := st.Status.Secret.Name
	if secretName == "" {
		secretName = "<pending>"
	}

	sections := []output.DescribeSection{
		{
			Fields: []output.DescribeField{
				{Key: "Name", Value: st.Name},
				{Key: "Namespace", Value: st.Namespace},
				{Key: "Created", Value: output.FormatAge(st.CreationTimestamp)},
			},
		},
		{
			Title: "Spec",
			Fields: []output.DescribeField{
				{Key: "Service Account", Value: orNone(st.Spec.ServiceAccountName)},
				{Key: "Input Resources", Value: inputResources},
			},
		},
		{
			Title: "Status",
			Fields: []output.DescribeField{
				{Key: "Secret", Value: secretName},
				{Key: "Description", Value: orNone(st.Status.FriendlyDescription)},
				{Key: "Observed Generation", Value: fmt.Sprintf("%d", st.Status.ObservedGeneration)},
			},
		},
	}

	if len(st.Status.Conditions) > 0 {
		sections = append(sections, output.DescribeSection{
			Title:  "Conditions",
			Fields: conditionFields(st.Status.Conditions),
		})
	}

	return p.PrintDescribe(sections)
}
