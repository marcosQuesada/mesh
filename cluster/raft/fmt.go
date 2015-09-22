package raft

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	"log"
	"time"
)

const (
	Election = message.Status("starting")
)

type Fsm struct {
	state string
}

type StateHandler func() StateHandler

func New() *Fsm {
	return &Fsm{
		state: "new",
	}
}

func (f *Fsm) Run() {
	for state := f.BootState; state != nil; {
		state = state()
	}

	fmt.Println("Done!")
}

func (f *Fsm) BootState() StateHandler {
	f.doThings("BootState")

	return f.FollowerState
}

func (f *Fsm) FollowerState() StateHandler {
	f.doThings("FollowerState")

	return f.CandidateState
}

func (f *Fsm) CandidateState() StateHandler {
	f.doThings("CandidateState")

	return f.LeaderState
}

func (f *Fsm) LeaderState() StateHandler {
	f.doThings("LeaderState")

	return nil
}

func (f *Fsm) doThings(st string) {
	log.Println(st)
	f.state = st
	time.Sleep(time.Second * 1)
}
