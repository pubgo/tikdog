package tikdog_job

type Job interface {
	Name() string
	Version() string
	Kind() string
	Type() string
	Close() error
	OnEvent(event interface{}) error
}
