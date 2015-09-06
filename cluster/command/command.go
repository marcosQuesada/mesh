package command

type Commandable interface {
	Execute(Command) Response
}

type Command struct {
	Name string
	Args Args
	Response Response
}

type Args []interface{}

type Response interface{}