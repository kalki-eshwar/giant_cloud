package cloud

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestRegisterRequestMarshalRoundTrip(t *testing.T) {
	in := &RegisterRequest{Address: "127.0.0.1:8080", Capacity: 512}

	encoded, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	out := &RegisterRequest{}
	if err := proto.Unmarshal(encoded, out); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !proto.Equal(in, out) {
		t.Fatalf("messages differ after round-trip: in=%+v out=%+v", in, out)
	}
}

func TestAllocateResponseMarshalRoundTrip(t *testing.T) {
	in := &AllocateResponse{NodeAddresses: []string{"node-a", "node-b", "node-c"}}

	encoded, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	out := &AllocateResponse{}
	if err := proto.Unmarshal(encoded, out); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(in.NodeAddresses, out.NodeAddresses) {
		t.Fatalf("node addresses differ after round-trip: in=%v out=%v", in.NodeAddresses, out.NodeAddresses)
	}
}
