package grpcclient

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/rpcpool/yellowstone-grpc/examples/golang/config"
	"github.com/rpcpool/yellowstone-grpc/examples/golang/processor"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second,
	Timeout:             time.Second,
	PermitWithoutStream: true,
}

// Connect initializes a gRPC connection
func Connect(address string, plaintext bool) *grpc.ClientConn {
	var opts []grpc.DialOption
	if plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, _ := x509.SystemCertPool()
		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	log.Println("Starting grpc client, connecting to", address)
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	return conn
}

func BuildSubscribeRequest(config config.Config) *pb.SubscribeRequest {
	subscription := &pb.SubscribeRequest{
		Transactions: make(map[string]*pb.SubscribeRequestFilterTransactions),
		Commitment:   pb.CommitmentLevel_PROCESSED.Enum(),
	}

	if config.Subscriptions.Transaction.Enable {
		subscription.Transactions["transactions_sub"] = &pb.SubscribeRequestFilterTransactions{
			Vote:           &config.Subscriptions.Transaction.TransactionsVote,
			Failed:         &config.Subscriptions.Transaction.TransactionsFailed,
			AccountInclude: config.Subscriptions.Transaction.TransactionsAccountsInclude,
		}
	}

	return subscription
}

// Subscribe handles the gRPC subscription logic
func Subscribe(conn *grpc.ClientConn, config config.Config) {
	client := pb.NewGeyserClient(conn)
	subscription := BuildSubscribeRequest(config)

	subscriptionJson, err := json.Marshal(&subscription)
	if err != nil {
		log.Printf("Failed to marshal subscription request: %v", subscriptionJson)
	}
	log.Printf("Subscription request: %s", string(subscriptionJson))

	// Set up the subscription request
	ctx := context.Background()
	if config.Token != "" {
		md := metadata.New(map[string]string{"x-token": config.Token})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	stream, err := client.Subscribe(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}

	err = stream.Send(subscription)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Create a set of owner addresses for quick lookup
	ownerSet := make(map[string]struct{})
	for _, owner := range config.Subscriptions.Transaction.TransactionsAccountsInclude {
		ownerSet[owner] = struct{}{}
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Fatalf("EOF received, exiting")
			return
		} else if err != nil {
			log.Fatalf("Error occurred in receiving update: %v", err)
		}

		go processor.HandleTransaction(resp, ownerSet)
	}
}
