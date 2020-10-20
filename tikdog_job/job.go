package tikdog_job

type Job interface {
	Name() string
	Version() string
	Kind() string
	Type() string
	OnEvent(event interface{}) error
}
