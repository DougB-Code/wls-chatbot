// provider_command.go defines AI CLI adapters for provider workflows.
// internal/ui/adapters/cli/ai/provider_command.go
package ai

import (
	"context"
	"fmt"
	"strings"

	providerfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/provider"
	"github.com/spf13/cobra"
)

// newProviderCommand creates the parent 'provider' command.
func newProviderCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage AI providers",
	}
	cmd.AddCommand(newProviderListCommand(deps))
	cmd.AddCommand(newProviderTestCommand(deps))
	cmd.AddCommand(newProviderAddCommand(deps))
	cmd.AddCommand(newProviderRemoveCommand(deps))
	cmd.AddCommand(newProviderCredentialsCommand(deps))
	return cmd
}

// newProviderListCommand lists configured providers.
func newProviderListCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured providers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			providers, err := applicationFacade.Providers.GetProviders(context.Background())
			if err != nil {
				return err
			}

			fmt.Printf("%-15s %-25s %-10s %-10s\n", "NAME", "DISPLAY NAME", "CONNECTED", "ACTIVE")
			fmt.Println(strings.Repeat("-", 70))
			for _, provider := range providers {
				connected := "no"
				active := "no"
				if provider.IsConnected {
					connected = "yes"
				}
				if provider.IsActive {
					active = "yes"
				}
				fmt.Printf("%-15s %-25s %-10s %-10s\n", provider.Name, provider.DisplayName, connected, active)
			}
			return nil
		},
	}
	return cmd
}

// newProviderTestCommand tests a provider connection.
func newProviderTestCommand(deps Dependencies) *cobra.Command {

	var providerName string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test a provider connection",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			fmt.Printf("Testing connection to %s...\n", providerName)
			if err := applicationFacade.Providers.TestProvider(context.Background(), providerName); err != nil {
				fmt.Printf("Connection failed: %v\n", err)
				return err
			}
			fmt.Println("Connection successful.")
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "name", "", "Provider name to test")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// newProviderAddCommand configures a provider using supplied credentials.
func newProviderAddCommand(deps Dependencies) *cobra.Command {

	var name string
	var credentials []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Configure credentials for a provider",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			credentialMap, err := parseKeyValuePairs(credentials)
			if err != nil {
				return err
			}

			_, err = applicationFacade.Providers.AddProvider(context.Background(), name, providerfeature.ProviderCredentials(credentialMap))
			if err != nil {
				return err
			}

			fmt.Printf("Provider %s configured.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Provider name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringArrayVar(&credentials, "credential", nil, "Provider credential as key=value (repeat flag)")

	return cmd
}

// newProviderRemoveCommand disconnects a configured provider.
func newProviderRemoveCommand(deps Dependencies) *cobra.Command {

	var name string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Disconnect a provider",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}
			if err := applicationFacade.Providers.RemoveProvider(context.Background(), name); err != nil {
				return err
			}

			fmt.Printf("Provider %s disconnected.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Provider name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// newProviderCredentialsCommand updates provider credentials.
func newProviderCredentialsCommand(deps Dependencies) *cobra.Command {

	var name string
	var credentials []string

	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Update provider credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			credentialMap, err := parseKeyValuePairs(credentials)
			if err != nil {
				return err
			}
			if len(credentialMap) == 0 {
				return fmt.Errorf("at least one --credential key=value is required")
			}

			if err := applicationFacade.Providers.UpdateProviderCredentials(context.Background(), name, providerfeature.ProviderCredentials(credentialMap)); err != nil {
				return err
			}

			fmt.Printf("Credentials updated for provider %s.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Provider name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringArrayVar(&credentials, "credential", nil, "Provider credential as key=value (repeat flag)")
	return cmd
}

// parseKeyValuePairs parses repeated key=value CLI flags into a map.
func parseKeyValuePairs(entries []string) (map[string]string, error) {

	if len(entries) == 0 {
		return nil, nil
	}

	parsed := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("invalid key=value entry: %s", entry)
		}
		trimmedKey := strings.TrimSpace(key)
		trimmedValue := strings.TrimSpace(value)
		if trimmedKey == "" || trimmedValue == "" {
			return nil, fmt.Errorf("invalid key=value entry: %s", entry)
		}
		parsed[trimmedKey] = trimmedValue
	}
	return parsed, nil
}
