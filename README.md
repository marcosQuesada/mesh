[![Build Status](https://travis-ci.org/marcosQuesada/mesh.svg?branch=develop)](https://travis-ci.org/marcosQuesada/mesh)

Mesh [WIP]
==========
 Cluster proof of concept using Raft quorum.
 
## Milestones
 * Communication between peers using message passing
 * Basic Raft leader implementation
 * Command Quorum execution
 * State Machine replication

### Boot Cluster definition (3 member nodes as default)
```bash
mesh -addr=127.0.0.1:12000 -cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002 -cli=1111
mesh -addr=127.0.0.1:12001 -cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002 -cli=1112
mesh -addr=127.0.0.1:12002 -cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002 -cli=1113
```
Where:
 * addr: Port where Mesh is listen on
 * cluster: cluster list definition separated by commas
