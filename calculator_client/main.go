package main

import (
	"context"
	"fmt"
	"github.com/ErFUN-KH/simple-grpc-project/calculatorpb"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	fmt.Println("Client is running...")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to server: %v", err)
	}
	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	//doSum(c)

	//doServerStreaming(c)

	doClientStreaming(c)
}

func doSum(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a sum Unary RPC")

	req := &calculatorpb.SumRequest{
		FirstNumber: 40,
		SecondUmber: 2,
	}

	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling sum RPC: %v", err)
	}

	log.Printf("Response from server: %v", res.SumResult)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a PrimeDecomposition server streaming RPC")

	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 12,
	}

	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling PrimeDecomposition RPC: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error while streaming PrimeDecomposition RPC: %v", err)
		}
		fmt.Println(res.PrimeFactor)
	}
}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a ComputeAverage client streaming RPC")

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling stream RPC: %v", err)
	}

	numbers := []int32{2, 5, 7, 9, 12, 57}

	for _, number := range numbers {
		fmt.Printf("Sending number: %v\n", number)
		err := stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: number,
		})
		if err != nil {
			log.Fatalf("Error while sending stream: %v", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response: %v", err)
	}

	fmt.Printf("The average is: %v\n", res.GetAverage())
}
