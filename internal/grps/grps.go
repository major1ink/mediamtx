package RMS

import (
	"context"
	"fmt"

	"github.com/bluenviron/mediamtx/internal/conf"
	pb "github.com/major1ink/repGrpc/pkg/repGrpc"
	"google.golang.org/grpc"
)
type GrpcClient struct {
	ctx context.Context
	Use bool
	Server string
	Client pb.RMSClient
	Conn *grpc.ClientConn
}
func CreateGrpcClient(ctx context.Context, cfg conf.GRPC) (GrpcClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.GrpcAddress, cfg.GrpcPort), grpc.WithInsecure())
	if err != nil {
		return GrpcClient{}, err
	}
	client := GrpcClient{
		ctx:    ctx,
		Use:    cfg.Use,
		Server: cfg.ServerName,
		Client: pb.NewRMSClient(conn),
		Conn:   conn,
	}
	return client, nil
}
func (c *GrpcClient) Post (attribute,query string) ( error){
	_,err := c.Client.Post(c.ctx, &pb.Insert{	Server: c.Server,	Attribute: attribute,	Query: query,})
	if err != nil {
		if err.Error() == "rpc error: code = Canceled desc = context canceled" {
		ctx := context.Background()
		_, err = c.Client.Post(ctx, &pb.Insert{Server: c.Server, Attribute: attribute, Query: query,})
		if err != nil {
			return err
		}
		return nil
		}
		return err
	}
	return nil

}
func (c *GrpcClient) Close () {
c.Conn.Close()
}

func (c *GrpcClient) Select (streamName, argument string) (*pb.AnswerSelect, error){
	var structSelect pb.Select
	switch argument {
	case "CodeMP":
		structSelect = pb.Select{
			Stream: streamName,
			Server: c.Server,
			CodeMP: true,
			StatusRecord: true,
		}
	case "CodeMP_Contract":
		structSelect = pb.Select{
			Stream: streamName,
			Server: c.Server,
			CodeMPContract: true,
		}
	case "MountPoint":
		structSelect = pb.Select{
			Stream: streamName,
			Server: c.Server,
			MountPoint: true,
		}
	case "StatusRecord":
		structSelect = pb.Select{
			Stream: streamName,
			Server: c.Server,
			StatusRecord: true,
		}
	}
	r,err := c.Client.Get(c.ctx, &structSelect)
	if err != nil {
		return nil, err
	}
	return r, nil

}