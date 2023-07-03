package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	agentV1 "grpc-reverse-test/gen/proto/agent/v1"
	"grpc-reverse-test/server/manage"
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

		manager := manage.NewAgentManager().Listen(listener).Serve()

		for {
			time.Sleep(3 * time.Second)
			fmt.Println("agent list:")
			for _, agent := range manager.List() {
				fmt.Printf("agent: %s \nlast seen: %v\n", agent.ID(), agent.LastSeen())
				agentInfo, err := agent.AgentInfo(context.Background(), &agentV1.AgentInfoRequest{})
				if err != nil {
					if status.Code(err) == codes.Unavailable {
						fmt.Printf("agent %s offline\n", agent.ID())
						agent.Disconnect()
						continue
					}
					fmt.Printf("get agent info error: %v\n", err)
					return err
				} else {
					fmt.Printf("%s\n", agentInfo.AgentID)
				}
			}
		}

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch

		return nil
	},
}
