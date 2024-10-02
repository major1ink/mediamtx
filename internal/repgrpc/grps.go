package repgrpc

import (
	"context"
	"fmt"

	"github.com/bluenviron/mediamtx/internal/conf"
	pb "github.com/major1ink/repGrpc/pkg/repGrpc"
	"google.golang.org/grpc"
)
type GrpcClient struct {
	Ctx context.Context
	Use bool
	Server string
	Client pb.RMSClient
	UseCodeMPAttribute bool
	Conn *grpc.ClientConn
}
func CreateGrpcClient(ctx context.Context, cfg conf.GRPC) (GrpcClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.GrpcAddress, cfg.GrpcPort), grpc.WithInsecure())
	if err != nil {
		return GrpcClient{}, err
	}
	client := GrpcClient{
		Ctx:    ctx,
		Use:    cfg.Use,
		UseCodeMPAttribute: cfg.UseCodeMPAttribute,
		Server: cfg.ServerName,
		Client: pb.NewRMSClient(conn),
		Conn:   conn,
	}
	return client, nil
}
func (c *GrpcClient) Put (streamName string, status *int32) (error){
 _,err := c.Client.Put(c.Ctx, &pb.Update{Server: c.Server,	Stream: streamName,	Status: status,})
 if err != nil {
	return  err
 }
 return  nil
}

func (c *GrpcClient) Post (attribute,query string) ( error){
	_,err := c.Client.Post(c.Ctx, &pb.Insert{	Server: c.Server,	Attribute: attribute,	Query: query,})
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
	r,err := c.Client.Get(c.Ctx, &structSelect)
	if err != nil {
		return nil, err
	}
	return r, nil

}