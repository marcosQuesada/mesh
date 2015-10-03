package raft

import (
	"fmt"
	"github.com/marcosQuesada/mesh/node"
	"log"
	"math"
	"math/rand"
	"reflect"
	"time"
)

const (
	PingIntervalBaseDuration = time.Second * 10
	MaxRandomDuration        = 10000
)

type raftAction interface {
	action() string
}

type Raft struct {
	node    node.Node
	mates   map[string]node.Node
	leader  node.Node
	sndChan chan interface{}
	rcvChan chan interface{}
	ready   chan bool
	timer   *raftTimer
}

type raftTimer struct {
	timerIn     chan time.Duration
	timerStop   chan bool
	timerSignal chan bool
}

type PoolResult map[string]string

type RaftRequest struct {
	ResponseChan chan interface{}
	Cmd          interface{}
}

type votationResult struct {
	expected  string
	responses PoolResult
}

type StateHandler func() StateHandler
type FSM struct {
	node  node.Node
	State interface{}
	exit  chan struct{}
}

func New(node node.Node, mates map[string]node.Node) *FSM {
	t := &raftTimer{
		timerIn:     make(chan time.Duration, 10),
		timerStop:   make(chan bool, 10),
		timerSignal: make(chan bool, 10),
	}
	r := &Raft{
		node:        node,
		mates:       mates,
		sndChan:     make(chan interface{}, 10),
		rcvChan:     make(chan interface{}, 10),
		ready:       make(chan bool, 0),
		timer:       t,
	}

	return &FSM{
		node:  node,
		State: r,
		exit:  make(chan struct{}),
	}
}

func (f *FSM) Run() {
	log.Println("Begin Raft Manager")

	r, ok := f.State.(*Raft)
	if !ok {
		log.Println("FMT state not found!")
		return
	}
	go r.runTimer()
	defer close(r.sndChan)
	defer close(r.rcvChan)
	defer close(r.ready)
	defer close(r.timer.timerIn)
	defer close(r.timer.timerSignal)

	for state := r.FollowerState; state != nil; {
		select {
		case <-f.exit:
			return
		default:
			state = state()
		}
	}

	log.Println("Done!")

}

func (f *FSM) Exit() {
	close(f.exit)
}

func (f *FSM) Ready() chan bool {
	e, ok := f.State.(*Raft)
	if !ok {
		log.Panic("FMT State is not Raft!")

		return nil
	}

	return e.ready
}

//on follower state expects pings from leader on t intervals
//on nil leader wait random time and switch candida
func (r *Raft) FollowerState() StateHandler {
	log.Println("On Follower State")
	if r.voidLeader() {
		tWait := getRandomDuration(r.node)
		log.Println("NO LEADER On Follower State, TWAIT IS ", tWait)
		r.timer.timerIn <- tWait
	}

	select {
	case msg := <-r.rcvChan:
		cmd, ok := msg.(RaftRequest)
		if ok {
			vr, ok := cmd.Cmd.(*VoteRequest)
			if ok {
				if r.voidLeader() {
					cmd.ResponseChan <- vr.Candidate.String()
				} else {
					cmd.ResponseChan <- r.leader.String()
				}
			}

			ping, castOk := cmd.Cmd.(*PingRequest)
			if castOk {
				r.setLeader(ping.Leader)
				log.Println("RAFT PING! from ", ping.Leader)
				r.timer.timerIn <- PingIntervalBaseDuration

				//@TODO: Required to solve transactions, until implement pure Messages
				cmd.ResponseChan <- r.leader.String()
			}
		}

		return r.FollowerState

	case <-r.timer.timerSignal:
		return r.CandidateState
	}
}

func (r *Raft) CandidateState() StateHandler {
	log.Println("On CandidateState")
	//Send Vote request
	r.sndChan <- &VoteRequest{r.node}
	r.timer.timerIn <- time.Second * 2
	select {
	case msg := <-r.rcvChan:
		r.timer.timerIn <- time.Hour
		switch v := msg.(type) {
		case RaftRequest:
			msg.(RaftRequest).ResponseChan <- r.node.String()

		case PoolResult:
			if r.evaluate(votationResult{expected: r.node.String(), responses: msg.(PoolResult)}) {
				return r.LeaderState
			}
		default:
			log.Println("Candidate command unknown", reflect.TypeOf(v).String())
		}

		return r.CandidateState

	case <-r.timer.timerSignal:
		log.Println("On Candidate Timeout")
		return r.FollowerState
	}

}

//leader pings followers on random time < Max time -10
func (r *Raft) LeaderState() StateHandler {
	log.Println("XXXXXXX On Leader State XXXXX")

	if r.voidLeader() {
		r.timer.timerIn <- time.Second * 2
		r.sndChan <- &PingRequest{r.node}
	}
	r.setLeader(r.node)

	select {
	case msg := <-r.rcvChan:
		switch v := msg.(type) {
		case RaftRequest:
			cmd, ok := msg.(RaftRequest)
			if ok {
				_, ok := cmd.Cmd.(*VoteRequest)
				if ok {
					log.Println("Voting to me", r.node.String())
					cmd.ResponseChan <- r.node.String()
				}

				ping, castOk := cmd.Cmd.(*PingRequest)
				if castOk {
					log.Println("PING! received from ", ping.Leader)
					r.timer.timerIn <- time.Hour

					return r.CandidateState
				}
			}
		case PoolResult:
			//@TODO: Only evaluate on VoteRequest
			if r.evaluate(votationResult{expected: r.node.String(), responses: msg.(PoolResult)}) {
				return r.LeaderState
			}
		default:
			fmt.Println("unknown", reflect.TypeOf(v).String())
		}

	case <-r.timer.timerSignal:
		r.sndChan <- &PingRequest{r.node}
		r.timer.timerIn <- time.Second * 2
	}

	return r.LeaderState
}

func (f *FSM) Request() chan interface{} {
	r, ok := f.State.(*Raft)
	if !ok {
		log.Println("FMT state not found!")
		return nil
	}
	return r.sndChan
}

func (f *FSM) Response() chan interface{} {
	r, ok := f.State.(*Raft)
	if !ok {
		log.Println("FMT state not found!")
		return nil
	}
	return r.rcvChan
}

func (r *Raft) runTimer() {
	var timeout *time.Timer = time.NewTimer(time.Hour)
	for {
		select {
		case tSize := <-r.timer.timerIn:
			timeout.Reset(tSize)
		case <-r.timer.timerStop:
			timeout.Stop()
		case <-timeout.C:
			r.timer.timerSignal <- true
		}
	}
}

func (r *Raft) voidLeader() bool {
	return r.leader == (node.Node{})
}

func (r *Raft) setLeader(n node.Node) {
	if r.voidLeader() {
		//send ready Signal
		r.ready <- true
	}
	r.leader = n

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
	reqMajority := int(math.Floor(float64(len(r.mates))/2)) + 1
	log.Println("FavVotes ", favVotes, "NegVotes", negVotes, "required", reqMajority)

	return favVotes >= reqMajority
}

func getRandomDuration(node node.Node) time.Duration {
	rand.Seed(time.Now().Unix() * int64(node.Port))
	rnd := rand.Intn(MaxRandomDuration)

	return time.Millisecond * time.Duration(rnd)
}
