package main

import (
	"context"
	"fmt"
	"github.com/ErFUN-KH/simple-grpc-project/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("Client is running...")

	// SSL config
	tls := false
	opts := grpc.WithInsecure()
	if tls {
		certFile := "../ssl/ca.crt"
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "api.example.com")
		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certifiate: %v", sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect to server: %v", err)
	}
	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	//doSum(c)

	//doServerStreaming(c)

	//doClientStreaming(c)

	//doBiDiStreaming(c)

	// Correct call
	//doSquareRoot(c, 10)
	// Error call
	//doSquareRoot(c, -3)

	doSumWithDeadLine(c, 5*time.Second)
	doSumWithDeadLine(c, 3*time.Second)
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

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a FindMaximum BiDi streaming RPC")

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while calling stream RPC: %v", err)
	}

	waitingForChannel := make(chan struct{})

	// send go routine
	go func() {
		numbers := []int32{2, 8, 1, 5, 37, 28, 42}

		for _, number := range numbers {
			err := stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			if err != nil {
				log.Fatalf("Error while sending stream: %v", err)
			}
			time.Sleep(1000 * time.Millisecond)
		}

		err := stream.CloseSend()
		if err != nil {
			log.Fatalf("Error while closing stream: %v", err)
		}
	}()

	// receive go routine
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receving stream: %v", err)
				break
			}

			fmt.Printf("New maximun is: %v\n", res.Maximum)
		}
		close(waitingForChannel)
	}()

	<-waitingForChannel
}

func doSquareRoot(c calculatorpb.CalculatorServiceClient, number int32) {
	fmt.Println("Starting to do a SquareRoot Unary RPC")

	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: number})
	if err != nil {
		resError, ok := status.FromError(err)
		if ok {
			// Actual error from gRPC (user error)
			fmt.Printf("Error message from server: %v\n", resError.Message())
			fmt.Println(resError.Code())
			if resError.Code() == codes.InvalidArgument {
				log.Fatalln("We probably sent a negative number")
			}
		} else {
			log.Fatalf("Big error calling SquareRoot: %v", err)
		}
	}
	fmt.Printf("Result of square root of %v: %v\n\n", number, res.NumberRoot)
}

func doSumWithDeadLine(c calculatorpb.CalculatorServiceClient, time time.Duration) {
	fmt.Println("Starting to do a SumWithDeadLine RPC")

	req := &calculatorpb.SumWithDeadLineRequest{
		FirstNumber: 40,
		SecondUmber: 2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time)
	defer cancel()

	res, err := c.SumWithDeadLine(ctx, req)
	if err != nil {
		resError, ok := status.FromError(err)
		if ok {
			if resError.Code() == codes.DeadlineExceeded {
				log.Fatalln("Timeout was hit! Deadline was exceeded")
			} else {
				log.Fatalf("Unexpected error: %v", err)
			}
		} else {
			log.Fatalf("Error while calling sum RPC: %v", err)
		}
	}

	log.Printf("Response from server: %v\n\n", res.SumResult)
}
