package main

import (
	"context"
	"testing"

	pb "github.com/example/proto/cloud"
)

func TestRegisterAddsNodeOnlyOnce(t *testing.T) {
	DB = nil
	ws := newWorkerServer()

	first, err := ws.Register(context.Background(), &pb.RegisterRequest{Address: "worker-a", Capacity: 100})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if !first.Success || first.NodeId != "worker-a" {
		t.Fatalf("unexpected register response: %+v", first)
	}

	_, err = ws.Register(context.Background(), &pb.RegisterRequest{Address: "worker-a", Capacity: 200})
	if err != nil {
		t.Fatalf("second Register failed: %v", err)
	}

	if len(ws.nodes) != 1 {
		t.Fatalf("expected exactly one node, got %d", len(ws.nodes))
	}
	if ws.nodes[0] != "worker-a" {
		t.Fatalf("unexpected node entry: %q", ws.nodes[0])
	}
}

func TestAllocateNodesRoundRobinAcrossActiveWorkers(t *testing.T) {
	ws := newWorkerServer()
	ws.nodes = []string{"n1", "n2", "n3"}
	ws.streams["n1"] = &streamWrapper{}
	ws.streams["n3"] = &streamWrapper{}

	cs := newClientServer(ws)

	res, err := cs.AllocateNodes(context.Background(), &pb.AllocateRequest{NumChunks: 5})
	if err != nil {
		t.Fatalf("AllocateNodes failed: %v", err)
	}

	expected := []string{"n1", "n3", "n1", "n3", "n1"}
	if len(res.NodeAddresses) != len(expected) {
		t.Fatalf("expected %d assignments, got %d", len(expected), len(res.NodeAddresses))
	}
	for i := range expected {
		if res.NodeAddresses[i] != expected[i] {
			t.Fatalf("assignment %d: expected %q, got %q", i, expected[i], res.NodeAddresses[i])
		}
	}

	next, err := cs.AllocateNodes(context.Background(), &pb.AllocateRequest{NumChunks: 1})
	if err != nil {
		t.Fatalf("second AllocateNodes failed: %v", err)
	}
	if len(next.NodeAddresses) != 1 || next.NodeAddresses[0] != "n3" {
		t.Fatalf("expected round-robin to continue at n3, got %+v", next.NodeAddresses)
	}
}

func TestAllocateNodesNoActiveWorkers(t *testing.T) {
	ws := newWorkerServer()
	ws.nodes = []string{"n1"}
	cs := newClientServer(ws)

	res, err := cs.AllocateNodes(context.Background(), &pb.AllocateRequest{NumChunks: 2})
	if err != nil {
		t.Fatalf("AllocateNodes failed: %v", err)
	}
	if len(res.NodeAddresses) != 0 {
		t.Fatalf("expected no assignments, got %+v", res.NodeAddresses)
	}
}

func TestUploadManifestReturnsStableID(t *testing.T) {
	DB = nil
	ws := newWorkerServer()
	cs := newClientServer(ws)

	res, err := cs.UploadManifest(context.Background(), &pb.ManifestRequest{FileName: "archive.tar", TotalChunks: 8})
	if err != nil {
		t.Fatalf("UploadManifest failed: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success response")
	}
	if res.ManifestId != "manifest-archive.tar" {
		t.Fatalf("unexpected manifest id: %q", res.ManifestId)
	}
}
