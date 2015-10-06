package raft

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/router/handler"
	"log"
	"math"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

const (
	BootMaxRandomDuration = 10000
	RaftHearBeatMaxDuration = 5000
	VoteRequestTimeOut = time.Second * 1
	FOLLOWER                 = "follower"
	CANDIDATE                = "candidate"
	LEADER                   = "leader"
)

type RaftAction interface {
	action() string
}

type Raft struct {
	node        node.Node
	mates       map[string]node.Node
	leader      node.Node
	sndChan     chan interface{}
	rcvChan     chan interface{}
	ready       chan node.Node
	timer       *raftTimer
	state       string
	currentTerm int
	termMutex   sync.Mutex
	booted bool
}

type raftTimer struct {
	timerIn     chan time.Duration
	timerStop   chan bool
	timerSignal chan bool
}

type PoolResult map[string]bool

type RaftRequest struct {
	ResponseChan chan interface{}
	Cmd          RaftAction
}

// FSM declaration
type StateHandler func() StateHandler
type FSM struct {
	node  node.Node
	State interface{}
	exit  chan struct{}
}

func New(localNode node.Node, mates map[string]node.Node) *FSM {
	internalTimer := &raftTimer{
		timerIn:     make(chan time.Duration, 10),
		timerStop:   make(chan bool, 10),
		timerSignal: make(chan bool, 10),
	}

	cleanMates := make(map[string]node.Node, len(mates)-1)
	for k, v := range mates {
		if k != localNode.String() {
			cleanMates[k] = v
		}
	}

	r := &Raft{
		node:    localNode,
		mates:   cleanMates,
		sndChan: make(chan interface{}, 10),
		rcvChan: make(chan interface{}, 10),
		ready:   make(chan node.Node, 0),
		timer:   internalTimer,
	}

	return &FSM{
		node:  localNode,
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

func (f *FSM) Ready() chan node.Node {
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
	if !r.booted {
		tWait := r.getRandomDuration(BootMaxRandomDuration)
		log.Println("Booting Raft, limit first follower state to random time ", tWait)
		r.timer.timerIn <- tWait
		r.booted = true
	}
	r.setState(FOLLOWER)

	//On Leader lost , wait random time an
	if r.voidLeader() {
		r.timer.timerIn <- r.getRandomDuration(RaftHearBeatMaxDuration)
	}

	select {
	case rcvMsg := <-r.rcvChan:
		switch v := rcvMsg.(type) {
		//On Vote Request my answer depends if i am following a leader
		case handler.Request:
			req := rcvMsg.(handler.Request)
			msg := req.Msg.(*message.RaftVoteRequest)

			//On nil leader vote true
			vote := r.voidLeader()
			log.Println("Follower: Vote Request Candiate", msg.Candidate, "vote", vote, "treqTerm", msg.Term, "current term", r.currentTerm)
			req.ResponseChan <- vote
			r.timer.timerIn <- RaftHearBeatMaxDuration * time.Millisecond
		//On Heartbeat received, Leader still alive, restart timeout
		case *message.RaftHeartBeatRequest:
			cmd := rcvMsg.(*message.RaftHeartBeatRequest)
			r.setLeader(cmd.Leader)
			r.timer.timerIn <- RaftHearBeatMaxDuration * time.Millisecond

		default:
			fmt.Println("Follower unknown", reflect.TypeOf(v).String())
		}

		return r.FollowerState

	case <-r.timer.timerSignal:
		log.Println("Follower state timeout, removing leader")
		r.leader = node.Node{}

		return r.CandidateState
	}
}

func (r *Raft) CandidateState() StateHandler {
	r.setState(CANDIDATE)

	r.termMutex.Lock()
	r.currentTerm++
	r.termMutex.Unlock()

	//Send Vote request & voteRequest tiemout
	r.sndChan <- r.voteRequest(r.node)
	r.timer.timerIn <- VoteRequestTimeOut

	select {
	case rcvMsg := <-r.rcvChan:
		switch v := rcvMsg.(type) {
		//On Vote Request my answer is that I'm Candidate to leader
		case handler.Request:
			req := rcvMsg.(handler.Request)
			msg := req.Msg.(*message.RaftVoteRequest)
			log.Println("Candiadate: Vote Request from:", msg.Candidate, "treqTerm", msg.Term, "current term", r.currentTerm, r.node.String())
			req.ResponseChan <- false

		//On Heartbeat received, do nothing! wait vote Request progress
		case *message.RaftHeartBeatRequest:
			cmd := rcvMsg.(*message.RaftHeartBeatRequest)
			log.Println("CANDIDATE PING! received from ", cmd.Leader)

		//On Pool Result evaluate!
		case PoolResult:
			r.timer.timerStop <- true
			if r.evaluate(rcvMsg.(PoolResult)) {
				return r.LeaderState
			}
		default:
			log.Println("Unexpected Candidate Message, type unknown", reflect.TypeOf(v).String())
		}

	case <-r.timer.timerSignal:
		log.Println("On Candidate Timeout")
	}
	return r.FollowerState
}

//leader pings followers on random time < Max time -10
func (r *Raft) LeaderState() StateHandler {
	r.setState(LEADER)

	if r.voidLeader() {
		//On success Election send HearBeat to mates
		r.sendHearBeat()

		r.timer.timerIn <- r.getRandomDuration(RaftHearBeatMaxDuration)
	}
	r.setLeader(r.node)

	select {
	case rcvMsg := <-r.rcvChan:
		switch v := rcvMsg.(type) {
		//On Vote Request my answer is that I'm the leader
		case handler.Request:
			req := rcvMsg.(handler.Request)
			msg := req.Msg.(*message.RaftVoteRequest)
			log.Println("Leader has received VoteRequest from:", msg.Candidate, "treqTerm", msg.Term, "current term", r.currentTerm)
			req.ResponseChan <- false

		//On Heartbeat received, maybe there's new leader, go candidate!
		case *message.RaftHeartBeatRequest:
			cmd := rcvMsg.(*message.RaftHeartBeatRequest)
			log.Println("Leader has received Heartbeat from ", cmd.Leader)
			r.timer.timerStop <- true

			return r.CandidateState

		default:
			fmt.Println("Leader unknown", reflect.TypeOf(v).String())
		}

	case <-r.timer.timerSignal:
		r.sendHearBeat()
		//restart random timer
		r.timer.timerIn <- time.Millisecond * 500 + r.getRandomDuration(RaftHearBeatMaxDuration - 500)
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

func (r *Raft) sendHearBeat() {
	for _, mate := range r.mates {
		go func(n node.Node) {
			r.sndChan <- &message.RaftHeartBeatRequest{Id: message.NewId(), From: r.node, To: n, Leader: r.node}
		}(mate)
	}
}

func (r *Raft) voteRequest(candidate node.Node) (msgs []message.Message) {
	r.termMutex.Lock()
	term := r.currentTerm
	r.termMutex.Unlock()

	for _, mate := range r.mates {
		msg := &message.RaftVoteRequest{
			Id:        message.NewId(),
			From:      r.node,
			To:        mate,
			Candidate: candidate,
			Term:      term,
/*			LastLogIndex ID
			LastLogTerm  ID*/
		}
		msgs = append(msgs, msg)
	}

	return
}

//run internal Timer
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
		r.ready <- n
	}
	r.leader = n

}

func (r *Raft) evaluate(vr PoolResult) bool {
	var favVotes, negVotes int
	//local Node is Candidate Request, always vote true!
	favVotes = 1
	for _, a := range vr {
		if a {
			favVotes++
			continue
		}
		negVotes++
	}

	//quorum is half part + 1 vote!
	reqMajority := int(math.Floor(float64(len(r.mates)+1)/2 + 0.5))
	if math.Mod(float64(len(r.mates)+1), 2) == 0 {
		reqMajority++
	}

	log.Println("FavVotes ", favVotes, "NegVotes", negVotes, "required", reqMajority)

	return favVotes >= reqMajority
}

func (r *Raft) setState(st string) {
	if r.state != st {
		log.Println("XXXX Raft State Changed from ", r.state, "to", st)
		r.state = st
	}
}

func (r *Raft) getRandomDuration(maxDuration int) time.Duration {
	rand.Seed(time.Now().Unix() * int64(r.node.Port))
	rnd := rand.Intn(maxDuration)

	return time.Millisecond * time.Duration(rnd)
}
