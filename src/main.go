package main

import ("ex_2_4/src/packages/peer")

func main() {
	var p = peer.Peer{}
	p.InitializePeer()
	p.StartPeer()
	for true {}
}