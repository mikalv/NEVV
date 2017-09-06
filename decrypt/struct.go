package decrypt

import (
	"github.com/qantik/nevv/api"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
)

// Name defines the protocol identifier in the onet service.
const Name = "decrypt"

// Prompt is the message sent from one node to another to invoke a new decryption
// of Box of ballots at the receiver.
type Prompt struct {
	Box *api.Box
}

// MessagePrompt wraps the Prompt message. For compatibilty reasons demanded
// by the onet framework.
type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

// Terminate is sent by the leaf node to the root node to signal that the last
// decryption of Box of shuffles has been perfomed.
type Terminate struct {
	Box *api.Box
}

// MessageTerminate wraps the Terminate message. For compatibility reasons
// demanded by the onet framework.
type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}

type Announce struct {
	Message string
}

// StructAnnounce just contains Announce and the data necessary to identify and
// process the message in the sda framework.
type StructAnnounce struct {
	*onet.TreeNode
	Announce
}

// Reply returns the count of all children.
type Reply struct {
	ChildrenCount int
	I             int
	Secret        abstract.Scalar
}

// StructReply just contains Reply and the data necessary to identify and
// process the message in the sda framework.
type StructReply struct {
	*onet.TreeNode
	Reply
}
