[![Build Status](https://travis-ci.org/marcosQuesada/mesh.svg?branch=develop)](https://travis-ci.org/marcosQuesada/mesh)

Mesh [WIP]
==========

 Proof of concept on how to build a basic cluster structure using Raft

Milestones
===========
-Basic Raft cluster
-Command Quorum execution

Cluster definition will be something like that
==============================================

mesh -addr=127.0.0.1:12000 cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002
mesh -addr=127.0.0.1:12001 cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002
mesh -addr=127.0.0.1:12002 cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002

Where:
	addr: Port where Mesh is listen on
	cluster: cluster list definition separated by commas



#NOTES
Node Name is Server:Port 
Communication between peers using message passing