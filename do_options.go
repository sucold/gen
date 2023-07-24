package gen

// DOOption gorm option interface
type DOOption[T any] interface {
	Apply(*DOConfig) error
	AfterInitialize(*DO[T]) error
}

type DOConfig struct {
}

// Apply update config to new config
func (c *DOConfig) Apply(config *DOConfig) error {
	if config != c {
		*config = *c
	}
	return nil
}

// AfterInitialize initialize plugins after db connected
func (c *DOConfig) AfterInitialize(db *DO[T]) error {
	return nil
}
