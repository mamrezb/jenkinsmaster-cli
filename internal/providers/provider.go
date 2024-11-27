package providers

type Provider interface {
	GetName() string
	Deploy() error
	RequiresTerraform() bool
}
