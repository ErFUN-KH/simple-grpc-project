package main

import (
	"context"
	"fmt"
	"github.com/ErFUN-KH/simple-grpc-project/calculatorpb"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
)

type server struct{}

func main() {
	fmt.Println("Server is running...")

	// Make a listener
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Make a gRPC server
	grpcServer := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(grpcServer, &server{})

	// Run the gRPC server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Received Sum RPC: %v\n", req)

	firstNumber := req.GetFirstNumber()
	secondNumber := req.GetSecondUmber()

	sum := firstNumber + secondNumber

	res := &calculatorpb.SumResponse{
		SumResult: sum,
	}

	return res, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Printf("Received PrimeNumberDecomposition RPC: %v\n", req)

	number := req.Number
	divisor := int64(2)

	for number > 1 {
		if number%divisor == 0 {
			err := stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})
			if err != nil {
				log.Fatalf("Failed to send response: %v\n", err)
			}

			number = number / divisor
		} else {
			divisor++
			fmt.Printf("Divisor has increased to %v", divisor)
		}
	}

	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("Received ComputeAverage RPC\n")

	sum := float64(0)
	count := float64(0)

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: sum / count,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		sum += float64(req.GetNumber())
		count++
	}
}
