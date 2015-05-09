package server

import "fmt"

//Mesh takes care on handling peers as unique members in swarm

type mesh struct {
	members map[address]bool
	peers   map[string]*Peer
}

type address string

func InitMesh(m []address) *mesh {
	fmt.Println("Mesh INIT m ", m, len(m))
	addresses := make(map[address]bool, len(m))
	for k, a := range m {
		fmt.Println("Mesh INIT a ", a, k)
		addresses[a] = false
	}
	fmt.Println("Mesh INIT ", addresses)

	return &mesh{
		members: addresses,
		peers:   make(map[string]*Peer, len(m)),
	}
}

func (m *mesh) JoinRequest(p *Peer) bool {
	fmt.Println("Join req from is ", p.Id().Address())
	if m.exist(p) {
		return false
	}
	m.members[address(p.Id().Address())] = true
	m.add(p)

	return true
}

func (m *mesh) Completed() bool {
	fmt.Println("Mesh is ", m.members)
	for _, m := range m.members {
		if !m {
			return false
		}
	}
	return true
}

func (m *mesh) add(p *Peer) {
	m.peers[p.Id().Address()] = p
}

func (m *mesh) remove(p *Peer) {
	delete(m.peers, p.Id().Address())
}

func (m *mesh) exist(p *Peer) (ok bool) {
	_, ok = m.peers[p.Id().Address()]

	return
}
