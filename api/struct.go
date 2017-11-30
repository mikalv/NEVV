package api

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/chains"
)

func init() {
	network.RegisterMessage(Ping{})
	network.RegisterMessage(Link{})
	network.RegisterMessage(LinkReply{})
	network.RegisterMessage(Open{})
	network.RegisterMessage(OpenReply{})
	network.RegisterMessage(Cast{})
	network.RegisterMessage(CastReply{})
	network.RegisterMessage(Finalize{})
	network.RegisterMessage(FinalizeReply{})
}

// Ping is the network probing message to check whether the service
// is contactable and running. In the successful case, another ping
// message is returned to the client.
type Ping struct {
	// Nonce is a random integer chosen by the client.
	Nonce uint32 `protobuf:"1,req,nonce"`
}

// Link is sent to the service whenever a new master skipchain has to be
// created. The message is only processed if the pin is corresponds to the
// pin of the running service. Since it is not part of the official API, its
// origin should be nevv's command line tool.
type Link struct {
	// Pin identifier of the service.
	Pin string `protobuf:"1,req,pin"`
	// Roster specifies the nodes handling the master skipchain.
	Roster *onet.Roster `protobuf:"2,opt,roster"`
	// Key is the frontend public key.
	Key abstract.Point `protobuf:"3,opt,key"`
	// Admins is a list of responsible admin (sciper numbers) users.
	Admins []chains.User `protobuf:"4,opt,admins"`
}

// LinkReply is returned when a master skipchain has been successfully created.
// It is only supposed to be sent to the command line tool since it is not part
// of the official API.
type LinkReply struct {
	// Master is the id of the genesis block of the master Skipchain.
	Master []byte `protobuf:"1,opt,master"`
}

type Login struct {
	Master    []byte      `protobuf:"1,req,master"`
	User      chains.User `protobuf:"2,req,sciper"`
	Signature []byte      `protobuf:"3,req,signature"`
}

type LoginReply struct {
	Token     string             `protobuf:"1,req,token"`
	Elections []*chains.Election `protobuf:"2,rep,elections"`
}

type Open struct {
	Token    string           `protobuf:"1,req,token"`
	Master   []byte           `protobuf:"2,req,master"`
	Election *chains.Election `protobuf:"2,req,election"`
}

type OpenReply struct {
	Genesis []byte         `protobuf:"1,req,genesis"`
	Key     abstract.Point `protobuf:"2,req,key"`
}

type Cast struct {
	Token   string  `protobuf:"1,req,token"`
	Genesis []byte  `protobuf:"2,req,genesis"`
	Ballot  *Ballot `protobuf:"3,req,ballot"`
}

type CastReply struct {
	Index uint32 `protobuf:"1,req,index"`
}

type Finalize struct {
	Token   string `protobuf:"1,req,token"`
	Genesis []byte `protobuf:"2,req,genesis"`
}

type FinalizeReply struct {
	Time uint32 `protobuf:"1,req,time"`
}
