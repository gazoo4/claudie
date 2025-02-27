package manifest

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validate validates the parsed manifest data.
func (m *Manifest) Validate() error {
	if err := validator.New().Struct(m); err != nil {
		return fmt.Errorf("failed to validate manifest: %w", err)
	}

	// check if at least one provider is defined
	// https://github.com/Berops/claudie/blob/master/docs/input-manifest/input-manifest.md#providers
	if len(m.Providers.GCP)+len(m.Providers.Hetzner) < 1 {
		return fmt.Errorf("need to define atleast one provider inside the providers section of the manifest")
	}

	if err := m.Providers.Validate(); err != nil {
		return fmt.Errorf("failed to validate providers section inside manifest: %w", err)
	}

	if err := m.NodePools.Validate(m); err != nil {
		return fmt.Errorf("failed to validate nodepools section inside manifest: %w", err)
	}

	if err := m.Kubernetes.Validate(m); err != nil {
		return fmt.Errorf("failed to validate kubernetes section inside manifest: %w", err)
	}

	if err := m.LoadBalancer.Validate(m); err != nil {
		return fmt.Errorf("failed to validate loadbalancers section inside manifest: %w", err)
	}

	return nil
}
