package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	pb "github.com/example/proto/cloud"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Warning: Redis ping failed (is Redis running?): %v", err)
	}
}

func startEventLoop(ws *workerServer) {
	// A simple loop polling a Redis List simulating a Stream for proofs
	ctx := context.Background()
	rrIndex := 0
	for {
		// BLPOP blocks until an element is available
		res, err := RedisClient.BLPop(ctx, 2*time.Second, "proof_tasks").Result()
		if err == redis.Nil {
			continue // timeout, retry
		} else if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// res[0] is key, res[1] is value
		// Expecting format: "node_address:job_id:chunk_id"
		// Simplified for MVP brevity
		taskStr := res[1]
		log.Printf("Redis Event: Scheduling proof challenge: %s", taskStr)
		
		// In a real app we'd parse the string to get node, job, chunk.
		// For now we push a dummy task to an available node using round-robin
		ws.mu.Lock()
		var activeNodes []string
		for _, n := range ws.nodes {
			if _, ok := ws.streams[n]; ok {
				activeNodes = append(activeNodes, n)
			}
		}
		ws.mu.Unlock()

		if len(activeNodes) > 0 {
			targetNode := activeNodes[rrIndex%len(activeNodes)]
			rrIndex++
			ws.pushTask(targetNode, "proof_challenge", "job-from-redis", "chunk-from-redis")
		} else {
			log.Printf("Redis Event: No active nodes to schedule proof challenge.")
		}
	}
}

func main() {
	os.Setenv("GODEBUG", "http2server=0,http2client=0")
	InitDB()
	InitRedis()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	
	ws := newWorkerServer()
	LoadNodes(ws)
	pb.RegisterWorkerServiceServer(grpcServer, ws)
	pb.RegisterClientServiceServer(grpcServer, newClientServer(ws))

	// Start the Redis event listener in the background
	go startEventLoop(ws)

	// Start the HTTP API server for the React GUI
	go startHTTPServer(ws)

	log.Printf("Control Plane gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
