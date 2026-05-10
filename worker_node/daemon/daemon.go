package daemon

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/example/proto/cloud"
)

var Client pb.WorkerServiceClient

func StartDaemon(address string, capacity int32) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// Note: normally we would defer conn.Close(), but daemon runs forever

	Client = pb.NewWorkerServiceClient(conn)

	ctx := context.Background()

	// Register
	regRes, err := Client.Register(ctx, &pb.RegisterRequest{
		Address:  address,
		Capacity: capacity,
	})
	if err != nil {
		log.Fatalf("could not register: %v", err)
	}
	log.Printf("Registered with Control Plane successfully: NodeID=%s", regRes.NodeId)

	// Heartbeat Loop
	go func() {
		for {
			stream, err := Client.Heartbeat(ctx)
			if err != nil {
				log.Printf("Failed to open heartbeat stream: %v, retrying...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Send initial heartbeat
			err = stream.Send(&pb.HeartbeatRequest{
				NodeId:   regRes.NodeId,
				Capacity: capacity,
			})
			if err != nil {
				log.Printf("Failed to send heartbeat: %v, reconnecting...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			log.Printf("Heartbeat stream connected.")

			// Start a goroutine to send periodic heartbeats
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				ticker := time.NewTicker(10 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						err := stream.Send(&pb.HeartbeatRequest{
							NodeId:   regRes.NodeId,
							Capacity: capacity,
						})
						if err != nil {
							log.Printf("Failed to send periodic heartbeat: %v", err)
							return
						}
					}
				}
			}()

			// Listen for push tasks
			for {
				res, err := stream.Recv()
				if err != nil {
					log.Printf("Heartbeat stream closed by server: %v", err)
					break // break inner loop to reconnect stream
				}
				log.Printf("Received push task from Control Plane: %s (Job: %s, Chunk: %s)", res.TaskType, res.JobId, res.ChunkId)

				if res.TaskType == "proof_challenge" {
					// Simulate executing a proof and responding
					go func(jobID, chunkID string) {
						time.Sleep(1 * time.Second) // Simulated proof work
						_, pErr := Client.SubmitProof(context.Background(), &pb.ProofRequest{
							NodeId:  regRes.NodeId,
							JobId:   jobID,
							ChunkId: chunkID,
							Hash:    "dummy_hash_12345",
						})
						if pErr != nil {
							log.Printf("Failed to submit proof for %s/%s: %v", jobID, chunkID, pErr)
						} else {
							log.Printf("Successfully submitted proof for %s/%s", jobID, chunkID)
						}
					}(res.JobId, res.ChunkId)
				}
			}
			cancel() // Stop the periodic heartbeat goroutine
			time.Sleep(5 * time.Second) // wait before reconnecting
		}
	}()
}
