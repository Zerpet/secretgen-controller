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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sigsyaml "sigs.k8s.io/yaml"
)

func newCertificateCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "certificate",
		Short:   "Manage Certificate resources",
		Aliases: []string{"cert", "certificates"},
	}
	cmd.AddCommand(
		newCertificateCreateCmd(opts),
		newCertificateUpdateCmd(opts),
		newCertificateDeleteCmd(opts),
		newCertificateGetCmd(opts),
		newCertificateDescribeCmd(opts),
		newCertificateListCmd(opts),
	)
	return cmd
}

// certFlags holds create/update flags for Certificate.
type certFlags struct {
	isCA               bool
	caRef              string
	commonName         string
	organization       string
	altNames           []string
	extKeyUsage        []string
	duration           int64
	secretTemplateFile string
}

func addCertFlags(cmd *cobra.Command, f *certFlags) {
	cmd.Flags().BoolVar(&f.isCA, "is-ca", false, "Whether this is a CA certificate")
	cmd.Flags().StringVar(&f.caRef, "ca-ref", "", "Name of CA Certificate to sign with")
	cmd.Flags().StringVar(&f.commonName, "common-name", "", "Certificate Common Name (CN)")
	cmd.Flags().StringVar(&f.organization, "organization", "", "Certificate Organization field")
	cmd.Flags().StringSliceVar(&f.altNames, "alt-names", nil, "Comma-separated Subject Alternative Names (IPs or DNS names)")
	cmd.Flags().StringSliceVar(&f.extKeyUsage, "ext-key-usage", nil, "Comma-separated Extended Key Usage values (client_auth, server_auth)")
	cmd.Flags().Int64Var(&f.duration, "duration", 0, "Certificate validity in days (0 = controller default of 365)")
	cmd.Flags().StringVar(&f.secretTemplateFile, "secret-template-file", "", "Path to YAML file defining a secretTemplate override")
}

func buildCertSpec(f *certFlags, cmd *cobra.Command) (sgv1alpha1.CertificateSpec, error) {
	spec := sgv1alpha1.CertificateSpec{}

	if cmd.Flags().Changed("is-ca") {
		spec.IsCA = f.isCA
	}
	if cmd.Flags().Changed("ca-ref") && f.caRef != "" {
		spec.CARef = &corev1.LocalObjectReference{Name: f.caRef}
	}
	if cmd.Flags().Changed("common-name") {
		spec.CommonName = f.commonName
	}
	if cmd.Flags().Changed("organization") {
		spec.Organization = f.organization
	}
	if cmd.Flags().Changed("alt-names") {
		spec.AlternativeNames = f.altNames
	}
	if cmd.Flags().Changed("ext-key-usage") {
		spec.ExtendedKeyUsage = f.extKeyUsage
	}
	if cmd.Flags().Changed("duration") {
		spec.Duration = f.duration
	}
	if f.secretTemplateFile != "" {
		tmpl, err := loadSecretTemplateFile(f.secretTemplateFile)
		if err != nil {
			return spec, err
		}
		spec.SecretTemplate = tmpl
	}
	return spec, nil
}

func loadSecretTemplateFile(path string) (*sgv1alpha1.SecretTemplate, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading secret-template-file: %w", err)
	}
	var tmpl sgv1alpha1.SecretTemplate
	if err := sigsyaml.Unmarshal(b, &tmpl); err != nil {
		return nil, fmt.Errorf("parsing secret-template-file: %w", err)
	}
	return &tmpl, nil
}

func newCertificateCreateCmd(opts *Options) *cobra.Command {
	f := &certFlags{}
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Certificate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			spec, err := buildCertSpec(f, cmd)
			if err != nil {
				return err
			}
			cert := &sgv1alpha1.Certificate{
				TypeMeta:   metav1.TypeMeta{APIVersion: "secretgen.k14s.io/v1alpha1", Kind: "Certificate"},
				ObjectMeta: metav1.ObjectMeta{Name: args[0], Namespace: opts.ResolvedNamespace()},
				Spec:       spec,
			}
			created, err := opts.SGClient().SecretgenV1alpha1().Certificates(opts.ResolvedNamespace()).Create(
				context.Background(), cert, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Certificate %q created in namespace %q\n", created.Name, created.Namespace)
			return nil
		},
	}
	addCertFlags(cmd, f)
	return cmd
}

func newCertificateUpdateCmd(opts *Options) *cobra.Command {
	f := &certFlags{}
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a Certificate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			client := opts.SGClient().SecretgenV1alpha1().Certificates(opts.ResolvedNamespace())
			existing, err := client.Get(context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("is-ca") {
				existing.Spec.IsCA = f.isCA
			}
			if cmd.Flags().Changed("ca-ref") {
				if f.caRef == "" {
					existing.Spec.CARef = nil
				} else {
					existing.Spec.CARef = &corev1.LocalObjectReference{Name: f.caRef}
				}
			}
			if cmd.Flags().Changed("common-name") {
				existing.Spec.CommonName = f.commonName
			}
			if cmd.Flags().Changed("organization") {
				existing.Spec.Organization = f.organization
			}
			if cmd.Flags().Changed("alt-names") {
				existing.Spec.AlternativeNames = f.altNames
			}
			if cmd.Flags().Changed("ext-key-usage") {
				existing.Spec.ExtendedKeyUsage = f.extKeyUsage
			}
			if cmd.Flags().Changed("duration") {
				existing.Spec.Duration = f.duration
			}
			if f.secretTemplateFile != "" {
				tmpl, err := loadSecretTemplateFile(f.secretTemplateFile)
				if err != nil {
					return err
				}
				existing.Spec.SecretTemplate = tmpl
			}

			updated, err := client.Update(context.Background(), existing, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Certificate %q updated\n", updated.Name)
			return nil
		},
	}
	addCertFlags(cmd, f)
	return cmd
}

func newCertificateDeleteCmd(opts *Options) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a Certificate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			name := args[0]
			if !yes {
				fmt.Printf("Delete Certificate %q in namespace %q? [y/N] ", name, opts.ResolvedNamespace())
				reader := bufio.NewReader(os.Stdin)
				ans, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					fmt.Println("Cancelled.")
					return nil
				}
			}
			err := opts.SGClient().SecretgenV1alpha1().Certificates(opts.ResolvedNamespace()).Delete(
				context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Certificate %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newCertificateGetCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get a Certificate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			cert, err := opts.SGClient().SecretgenV1alpha1().Certificates(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			p := opts.Printer()
			if p.IsTable() {
				return p.PrintTable(
					[]string{"NAME", "IS-CA", "DESCRIPTION", "AGE"},
					[][]string{certTableRow(cert)},
				)
			}
			return p.PrintObject(cert)
		},
	}
}

func newCertificateDescribeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Certificate resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			cert, err := opts.SGClient().SecretgenV1alpha1().Certificates(opts.ResolvedNamespace()).Get(
				context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			return describeCertificate(opts.Printer(), cert)
		},
	}
}

func newCertificateListCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Certificate resources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Connect(); err != nil {
				return err
			}
			list, err := opts.SGClient().SecretgenV1alpha1().Certificates(opts.ResolvedNamespace()).List(
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
				rows[i] = certTableRow(&list.Items[i])
			}
			return p.PrintTable([]string{"NAME", "IS-CA", "DESCRIPTION", "AGE"}, rows)
		},
	}
}

func certTableRow(c *sgv1alpha1.Certificate) []string {
	return []string{
		c.Name,
		output.BoolToStr(c.Spec.IsCA),
		c.Status.FriendlyDescription,
		output.FormatAge(c.CreationTimestamp),
	}
}

func describeCertificate(p *output.Printer, c *sgv1alpha1.Certificate) error {
	caRef := "<none>"
	if c.Spec.CARef != nil {
		caRef = c.Spec.CARef.Name
	}
	dur := "365 (default)"
	if c.Spec.Duration != 0 {
		dur = fmt.Sprintf("%d", c.Spec.Duration)
	}

	sections := []output.DescribeSection{
		{
			Fields: []output.DescribeField{
				{Key: "Name", Value: c.Name},
				{Key: "Namespace", Value: c.Namespace},
				{Key: "Created", Value: output.FormatAge(c.CreationTimestamp)},
			},
		},
		{
			Title: "Spec",
			Fields: []output.DescribeField{
				{Key: "Is CA", Value: output.BoolToStr(c.Spec.IsCA)},
				{Key: "CA Ref", Value: caRef},
				{Key: "Common Name", Value: orNone(c.Spec.CommonName)},
				{Key: "Organization", Value: orNone(c.Spec.Organization)},
				{Key: "Alternative Names", Value: output.JoinStrings(c.Spec.AlternativeNames, ", ")},
				{Key: "Extended Key Usage", Value: output.JoinStrings(c.Spec.ExtendedKeyUsage, ", ")},
				{Key: "Duration (days)", Value: dur},
			},
		},
		{
			Title: "Status",
			Fields: []output.DescribeField{
				{Key: "Description", Value: orNone(c.Status.FriendlyDescription)},
				{Key: "Observed Generation", Value: fmt.Sprintf("%d", c.Status.ObservedGeneration)},
			},
		},
	}

	if len(c.Status.Conditions) > 0 {
		sections = append(sections, output.DescribeSection{
			Title: "Conditions",
			Fields: conditionFields(c.Status.Conditions),
		})
	}

	return p.PrintDescribe(sections)
}

func conditionFields(conds []sgv1alpha1.Condition) []output.DescribeField {
	fields := make([]output.DescribeField, len(conds))
	for i, cond := range conds {
		val := string(cond.Status)
		if cond.Message != "" {
			val += " — " + cond.Message
		}
		fields[i] = output.DescribeField{Key: string(cond.Type), Value: val}
	}
	return fields
}

func orNone(s string) string {
	if s == "" {
		return "<none>"
	}
	return s
}
