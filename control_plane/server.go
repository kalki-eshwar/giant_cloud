package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	pb "github.com/example/proto/cloud"
)

type streamWrapper struct {
	stream pb.WorkerService_HeartbeatServer
	mu     sync.Mutex
}

type workerServer struct {
	pb.UnimplementedWorkerServiceServer
	mu       sync.Mutex
	nodes    []string
	streams  map[string]*streamWrapper
}

func newWorkerServer() *workerServer {
	return &workerServer{
		streams: make(map[string]*streamWrapper),
	}
}

func (s *workerServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Worker registered: %s with capacity %d", req.Address, req.Capacity)
	s.mu.Lock()
	
	// Add to memory list for quick round-robin
	found := false
	for _, n := range s.nodes {
		if n == req.Address {
			found = true
			break
		}
	}
	if !found {
		s.nodes = append(s.nodes, req.Address)
	}
	s.mu.Unlock()

	// Persist to DB
	if DB != nil {
		_, err := DB.Exec("INSERT INTO nodes (address, capacity) VALUES ($1, $2) ON CONFLICT (address) DO UPDATE SET capacity = $2, status = 'active'", req.Address, req.Capacity)
		if err != nil {
			log.Printf("Failed to insert node to DB: %v", err)
		}
	}

	return &pb.RegisterResponse{Success: true, NodeId: req.Address}, nil
}

func (s *workerServer) Heartbeat(stream pb.WorkerService_HeartbeatServer) error {
	var nodeID string
	for {
		req, err := stream.Recv()
		if err != nil {
			if nodeID != "" {
				log.Printf("Worker disconnected: %s", nodeID)
				s.mu.Lock()
				// Only delete if the current stream matches the one in the map
				if sw, ok := s.streams[nodeID]; ok && sw.stream == stream {
					delete(s.streams, nodeID)
				}
				s.mu.Unlock()
			}
			return err
		}
		
		if nodeID == "" {
			nodeID = req.NodeId
			s.mu.Lock()
			s.streams[nodeID] = &streamWrapper{stream: stream}
			s.mu.Unlock()
			log.Printf("Heartbeat stream started for: %s", nodeID)
		}
	}
}

func (s *workerServer) SubmitProof(ctx context.Context, req *pb.ProofRequest) (*pb.ProofResponse, error) {
	log.Printf("Received proof from %s for job %s, chunk %s: %s", req.NodeId, req.JobId, req.ChunkId, req.Hash)
	return &pb.ProofResponse{Success: true}, nil
}

func (s *workerServer) pushTask(nodeID string, taskType string, jobID string, chunkID string) {
	s.mu.Lock()
	sw, ok := s.streams[nodeID]
	s.mu.Unlock()

	if ok {
		sw.mu.Lock()
		err := sw.stream.Send(&pb.HeartbeatResponse{
			TaskType: taskType,
			JobId:    jobID,
			ChunkId:  chunkID,
		})
		sw.mu.Unlock()
		if err != nil {
			log.Printf("Failed to push task to %s: %v", nodeID, err)
		}
	} else {
		log.Printf("Cannot push task, worker stream not found for %s", nodeID)
	}
}

type clientServer struct {
	pb.UnimplementedClientServiceServer
	workerSrv *workerServer
	rrIndex   uint64
	mu        sync.Mutex
}

func newClientServer(ws *workerServer) *clientServer {
	return &clientServer{workerSrv: ws}
}

func (s *clientServer) UploadManifest(ctx context.Context, req *pb.ManifestRequest) (*pb.ManifestResponse, error) {
	log.Printf("Client uploaded manifest for %s with %d chunks", req.FileName, req.TotalChunks)
	
	if DB != nil {
		_, err := DB.Exec("INSERT INTO manifests (file_name, total_chunks) VALUES ($1, $2)", req.FileName, req.TotalChunks)
		if err != nil {
			log.Printf("Failed to insert manifest to DB: %v", err)
		}
	}

	return &pb.ManifestResponse{Success: true, ManifestId: "manifest-" + req.FileName}, nil
}

func (s *clientServer) AllocateNodes(ctx context.Context, req *pb.AllocateRequest) (*pb.AllocateResponse, error) {
	s.workerSrv.mu.Lock()
	var activeNodes []string
	for _, n := range s.workerSrv.nodes {
		if _, ok := s.workerSrv.streams[n]; ok {
			activeNodes = append(activeNodes, n)
		}
	}
	s.workerSrv.mu.Unlock()

	nodesCount := uint64(len(activeNodes))
	if nodesCount == 0 {
		return &pb.AllocateResponse{}, nil
	}

	assigned := make([]string, req.NumChunks)
	s.mu.Lock()
	for i := 0; i < int(req.NumChunks); i++ {
		assigned[i] = activeNodes[s.rrIndex%nodesCount]
		s.rrIndex++
	}
	s.mu.Unlock()

	return &pb.AllocateResponse{NodeAddresses: assigned}, nil
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type NodeStatus struct {
	ID       string `json:"id"`
	Capacity int32  `json:"capacity"`
	Status   string `json:"status"` // "active" or "offline"
}

func startHTTPServer(ws *workerServer) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/nodes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		var nodesList []NodeStatus
		if DB != nil {
			rows, err := DB.Query("SELECT address, capacity, status FROM nodes")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var ns NodeStatus
					if err := rows.Scan(&ns.ID, &ns.Capacity, &ns.Status); err == nil {
						// Check if stream is actually active right now
						ws.mu.Lock()
						_, ok := ws.streams[ns.ID]
						ws.mu.Unlock()
						if ok {
							ns.Status = "active"
						} else {
							ns.Status = "offline"
						}
						nodesList = append(nodesList, ns)
					}
				}
			}
		} else {
			ws.mu.Lock()
			for _, n := range ws.nodes {
				_, ok := ws.streams[n]
				status := "offline"
				if ok {
					status = "active"
				}
				nodesList = append(nodesList, NodeStatus{ID: n, Capacity: 100, Status: status})
			}
			ws.mu.Unlock()
		}

		if nodesList == nil {
			nodesList = []NodeStatus{}
		}

		json.NewEncoder(w).Encode(nodesList)
	})

	mux.HandleFunc("/api/nodes/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Expecting /api/nodes/{id}/ping
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) != 5 || pathParts[4] != "ping" || r.Method != "POST" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
			return
		}
		nodeID := pathParts[3]

		ws.mu.Lock()
		_, ok := ws.streams[nodeID]
		ws.mu.Unlock()

		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Node is offline"})
			return
		}

		// Push task
		ws.pushTask(nodeID, "proof_challenge", "manual-job", "manual-chunk")

		json.NewEncoder(w).Encode(map[string]string{"status": "ping queued for " + nodeID})
	})

	log.Println("Control Plane HTTP server listening on :8081")
	if err := http.ListenAndServe(":8081", enableCORS(mux)); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
