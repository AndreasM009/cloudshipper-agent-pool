package repositories

// DefinitionRepository to access yaml definition and parameters of a tenant's deployment
type DefinitionRepository interface {
	Add(deploymentID string, yaml []byte, parameters []byte) error
	Get(deploymentID string) ([]byte, []byte, error)
	Delete(deploymentID string) error
}
