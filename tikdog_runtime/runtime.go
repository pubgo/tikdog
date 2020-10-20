package tikdog_runtime

type Runtime interface {
	Type() string
	Exec() error
}
