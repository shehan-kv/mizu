package common

type IDGenerator interface {
	Generate() (string, error)
}
