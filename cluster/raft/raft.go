package raft

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"log"
	"math"
	"math/rand"
	"time"
)

const (
	NodeStatus                    = message.Status("starting")
	PingIntervalBaseDuration      = time.Second * 10
	MaxRandomDuration             = 5000
	PingIntervalMaxRandomDuration = time.Millisecond * MaxRandomDuration
)

type raftAction interface {
	action() string
}

type Request struct {
	destination node.Node
	command     raftAction
}

type Raft struct {
	node         node.Node
	mates        []node.Node
	leader       node.Node
	request      chan interface{}
	response     chan interface{}
	ready        chan bool
	pingInterval time.Duration
	state        string
}

type StateHandler func() StateHandler

type votationResult struct {
	expected  string
	responses []string
}

type FSM struct {
	State    interface{}
	request  chan interface{}
	response chan interface{}
}

func New(node node.Node, mates []node.Node) *FSM {
	request := make(chan interface{}, 10)
	response := make(chan interface{}, 10)

	r := &Raft{
		node:         node,
		mates:        mates,
		state:        "new",
		request:      request,
		response:     response,
		pingInterval: PingIntervalBaseDuration + getRandomDuration(),
		ready:        make(chan bool, 0),
	}

	log.Println("Booting Raft FSM MATESSSSSSSSSSSSSSS", mates)

	return &FSM{State: r, request: request, response: response}
}

func (f *FSM) Run() {
	e, ok := f.State.(*Raft)
	if !ok {
		log.Println("FMT state not found!")
		return
	}
	log.Println("-----------------------------------------------------------BEGIN-----------------------------------------------------------")

	for state := e.FollowerState; state != nil; {
		state = state()
	}

	fmt.Println("Done!")

}

func (f *FSM) Ready() chan bool {
	e, ok := f.State.(*Raft)
	if !ok {
		log.Panic("FMT Ready state not found!")

		return nil
	}

	return e.ready
}

//on follower state expects pings from leader on t intervals
//on nil leader wait random time and switch candida
func (r *Raft) FollowerState() StateHandler {
	log.Println("----------------------------------------------------------On Follower State")
	if r.leader == (node.Node{}) {
		time.Sleep(getRandomDuration())
		log.Println("NO LEADER On Follower State")

		return r.CandidateState
	}

	timeout := time.NewTimer(PingIntervalMaxRandomDuration)
		select {
		case msg := <-r.response:
			timeout.Reset(r.pingInterval)
			log.Println("MSG RESPONSE ", msg)
			return r.FollowerState

		case <-timeout.C:
			log.Println("MSG TIMEOUT RESPONSE ")
			return r.CandidateState
		}
}

func (r *Raft) CandidateState() StateHandler {
	log.Println("On Candidate State")

	//Send Vote request
	for _, n := range r.mates {
		log.Println("Vote Request to ", n)
		r.request <- Request{n, &VoteRequest{r.node}}
	}

	timeout := time.NewTimer(time.Second * 2)
	select {
	case msg := <-r.response:
		timeout.Reset(r.pingInterval)
		log.Println("MSG CANDIDATE RESPONSE ", msg)

		//evaluate response
		//return r.LeaderState
		return r.CandidateState
	case <-timeout.C:
		log.Println("On Candidate Timeout")
		return r.FollowerState
	}

}

//leader pings followers on random time < Max time -10
func (r *Raft) LeaderState() StateHandler {
	log.Println("On Leader State")
	ping := time.NewTicker(r.pingInterval)
	/*	for {*/
	select {
	case msg := <-r.response:
		ping.Stop()

		log.Println("MSG RESPONSE ", msg)

		return r.FollowerState
	case <-ping.C:
		for _, n := range r.mates {
			log.Println("ping ", n)

			r.request <- Request{n, &PingRequest{}}
		}
	}
	/*	}*/

	return nil
}

func (r *Raft) evaluate(vr votationResult) bool {
	var favVotes, negVotes int
	for _, a := range vr.responses {
		if a == vr.expected {
			favVotes++
			continue
		}
		negVotes++
	}
	reqMajority := int(math.Floor(float64(len(vr.responses)+1)/2)) + 1
	log.Println("FavVotes ", favVotes, "NegVotes", negVotes, "required", reqMajority)

	return favVotes >= reqMajority
}

func (f *FSM) Request() chan interface{} {
	return f.request
}

func (f *FSM) Response() chan interface{} {
	return f.response
}

// Request Types
type VoteRequest struct {
	Candidate node.Node
}

func (v *VoteRequest) action() string {
	return "voteRequest"
}

type PingRequest struct {
}

func (v *PingRequest) action() string {
	return "PingRequest"
}

func getRandomDuration() time.Duration {
	rand.Seed(time.Now().Unix())

	return time.Millisecond * time.Duration(rand.Intn(MaxRandomDuration))
}
