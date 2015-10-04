package raft

import (
	"fmt"
	"github.com/marcosQuesada/mesh/node"
	"log"
	"math"
	"math/rand"
	"reflect"
	"time"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/router/handler"
)

const (
	PingIntervalBaseDuration = time.Second * 10
	MaxRandomDuration        = 10000
	FOLLOWER = "follower"
	CANDIDATE = "candidate"
	LEADER = "leader"
)

type RaftAction interface {
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
	state   string
}

type raftTimer struct {
	timerIn     chan time.Duration
	timerStop   chan bool
	timerSignal chan bool
}

type PoolResult map[string]string

type RaftRequest struct {
	ResponseChan chan interface{}
	Cmd          RaftAction
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

func New(localNode node.Node, mates map[string]node.Node) *FSM {
	t := &raftTimer{
		timerIn:     make(chan time.Duration, 10),
		timerStop:   make(chan bool, 10),
		timerSignal: make(chan bool, 10),
	}
	//delete(mates, node.String())
	newMates := make(map[string]node.Node, len(mates) -1)
	for k, v := range mates {
		if k != localNode.String() {
			newMates[k] = v
		}
	}
	r := &Raft{
		node:        localNode,
		mates:       newMates,
		sndChan:     make(chan interface{}, 10),
		rcvChan:     make(chan interface{}, 10),
		ready:       make(chan bool, 0),
		timer:       t,
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
	r.setState(FOLLOWER)

	//When boot enable timer to jump to candidate
	if r.voidLeader() {
		tWait := getRandomDuration(r.node)
		log.Println("NO LEADER On Follower State, TWAIT IS ", tWait)
		r.timer.timerIn <- tWait
	}

	select {
	case rcvMsg := <-r.rcvChan:
		switch v := rcvMsg.(type) {
		//On Vote Request my answer depends if i am following a leader
		case handler.Request:
			req := rcvMsg.(handler.Request)
			msg := req.Msg.(*message.RaftVoteRequest)
			vote := r.leader.String()
			if r.voidLeader() {
				vote = msg.Candidate.String()
			}
			log.Println("Follower: Vote Request Vote for myself", vote)
			req.ResponseChan <- vote

		//On Heartbeat received, Leader still alive, restart timeout
		case *message.RaftHeartBeatRequest:
			cmd := rcvMsg.(*message.RaftHeartBeatRequest)
			r.setLeader(cmd.Leader)
			r.timer.timerIn <- PingIntervalBaseDuration

		default:
			fmt.Println("unknown", reflect.TypeOf(v).String())
		}

		return r.FollowerState

	case <-r.timer.timerSignal:
		return r.CandidateState
	}
}

func (r *Raft) CandidateState() StateHandler {
	r.setState(CANDIDATE)

	//Send Vote request
	r.sndChan <-r.voteRequest(r.node)
	r.timer.timerIn <- time.Second * 2
	select {
	case rcvMsg := <-r.rcvChan:
		switch v := rcvMsg.(type) {

		//On Vote Request my answer is that I'm Candidate to leader
		case handler.Request:
			req := rcvMsg.(handler.Request)
			msg := req.Msg.(*message.RaftVoteRequest)
			log.Println("Candiadate: Vote Request from:", msg.Candidate ,"Vote for myself", r.node.String())
			req.ResponseChan <- r.node.String()

		//On Heartbeat received, do nothing! wait vote Request progress
		case *message.RaftHeartBeatRequest:
			cmd := rcvMsg.(*message.RaftHeartBeatRequest)
			log.Println("CANDIDATE PING! received from ", cmd.Leader)

		//On Pool Result evaluate!
		case PoolResult:
			r.timer.timerIn <- time.Hour
			if r.evaluate(votationResult{expected: r.node.String(), responses: rcvMsg.(PoolResult)}) {
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
		r.timer.timerIn <- time.Second * 2
		r.sendHearBeat()
	}
	r.setLeader(r.node)

	select {
	case rcvMsg := <-r.rcvChan:
		switch v := rcvMsg.(type) {
		//On Vote Request my answer is that I'm the leader
		case handler.Request:
			req := rcvMsg.(handler.Request)
			msg := req.Msg.(*message.RaftVoteRequest)
			log.Println("Leader has received VoteRequest from:", msg.Candidate ,"Vote for myself", r.node.String())
			req.ResponseChan <- r.node.String()


		//On Heartbeat received, maybe there's new leader, go candidate!
		case *message.RaftHeartBeatRequest:
			cmd := rcvMsg.(*message.RaftHeartBeatRequest)
			log.Println("Leader has received Heartbeat from ", cmd.Leader)
			r.timer.timerIn <- time.Hour

			return r.CandidateState

		default:
			fmt.Println("unknown", reflect.TypeOf(v).String())
		}

	case <-r.timer.timerSignal:
		r.sendHearBeat()
		r.timer.timerIn <- time.Second * 5 //PingIntervalBaseDuration
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
	for _, mate:= range r.mates {
		go func(n node.Node){
			r.sndChan <- &message.RaftHeartBeatRequest{Id: message.NewId(), From: r.node, To: n, Leader: r.node}
		}(mate)
	}
}

func (r *Raft) voteRequest(candidate node.Node) (msgs []message.Message) {
	for _, node := range r.mates {
		msg := &message.RaftVoteRequest{Id: message.NewId(), From: r.node, To: node, Candidate: candidate}
		msgs = append(msgs, msg)
	}

	return
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

func (r *Raft) setState(st string) {
	if r.state != st {
		log.Println("XXXX Raft State Changed from ", r.state, "to", st)
		r.state = st
	}
}

func getRandomDuration(node node.Node) time.Duration {
	rand.Seed(time.Now().Unix() * int64(node.Port))
	rnd := rand.Intn(MaxRandomDuration)

	return time.Millisecond * time.Duration(rnd)
}

