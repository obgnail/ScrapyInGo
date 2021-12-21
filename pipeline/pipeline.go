package pipeline

type Pipeline interface {
	ProcessItem(item interface{}) error
}
