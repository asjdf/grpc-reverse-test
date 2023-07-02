package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	agentV1 "grpc-reverse-test/gen/proto/agent/v1"
	"grpc-reverse-test/pkg/grpcTun"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server此时是调用服务的client端
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a server",
	RunE: func(cmd *cobra.Command, args []string) error {
		listener, err := net.Listen("tcp", ":5325")
		if err != nil {
			return err
		}
		s := grpcTun.NewServer().Listen(listener)
		s.OnConnect(func(conn net.Conn) error {
			grpcConn, err := grpc.Dial("", grpc.WithInsecure(),
				grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
					return conn, nil
				}),
			)
			if err != nil {
				return err
			}
			client := agentV1.NewDistributeServiceClient(grpcConn)
			for {
				agentInfo, err := client.AgentInfo(context.Background(), &agentV1.AgentInfoRequest{})
				if err != nil {
					fmt.Printf("get agent info error: %v\n", err)
					return err
				} else {
					fmt.Printf("get agent info: id = %s\n", agentInfo.AgentID)
				}
				time.Sleep(5 * time.Second)
			}
		})
		err = s.Init()
		if err != nil {
			return err
		}

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch

		return nil
	},
}
