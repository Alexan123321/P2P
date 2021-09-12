/**
BY: Deyana Atanasova, Henrik Tambo Buhl & Alexander Stæhr Johansen 
DATE: 12-09-2021 (revised)
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed chat implemented as structured P2P flooding network.
**/

package main

import ("ex_2_4/src/packages/peer")

func main() {
	var p = peer.Peer{}
	p.InitializePeer()
	p.StartPeer()
	for true {}
}
