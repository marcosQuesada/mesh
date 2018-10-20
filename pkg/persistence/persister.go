package persistence

import "github.com/marcosQuesada/mesh/pkg/cluster/command"

type Persister interface {
	Set(command.Args) error
	Get(command.Args) (command.Response, error)
}
