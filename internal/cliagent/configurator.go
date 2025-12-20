package cliagent

// Configurator is an optional interface that agents can implement to provide
// project-level setup and configuration. Agents implementing this interface
// can configure settings files, permissions, and other project-specific
// resources during autospec init.
type Configurator interface {
	// ConfigureProject sets up the agent for a specific project.
	// It configures settings files, permissions, and other project-specific
	// resources. The method should be idempotent - calling it multiple times
	// with the same arguments should produce the same result.
	//
	// Parameters:
	//   - projectDir: The root directory of the project
	//   - specsDir: The directory where specs are stored (e.g., "specs" or "features")
	//
	// Returns:
	//   - ConfigResult describing what was configured
	//   - error if configuration failed
	ConfigureProject(projectDir, specsDir string) (ConfigResult, error)
}

// ConfigResult describes the outcome of agent project configuration.
type ConfigResult struct {
	// PermissionsAdded lists permissions that were added during configuration.
	// Empty if no permissions were added (e.g., already configured).
	PermissionsAdded []string

	// AlreadyConfigured is true if all required configuration was already present.
	// When true, PermissionsAdded will typically be empty.
	AlreadyConfigured bool

	// Warning contains an optional warning message, such as when a deny list
	// entry conflicts with required permissions.
	Warning string
}

// Configure checks if the given agent implements Configurator and calls
// ConfigureProject if it does. Returns nil, nil if the agent does not
// implement Configurator.
func Configure(agent Agent, projectDir, specsDir string) (*ConfigResult, error) {
	configurator, ok := agent.(Configurator)
	if !ok {
		return nil, nil
	}

	result, err := configurator.ConfigureProject(projectDir, specsDir)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// IsConfigurator returns true if the given agent implements the Configurator
// interface.
func IsConfigurator(agent Agent) bool {
	_, ok := agent.(Configurator)
	return ok
}
