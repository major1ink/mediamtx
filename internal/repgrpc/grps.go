package repgrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/bluenviron/mediamtx/internal/conf"
	pb "github.com/major1ink/repGrpc/pkg/repGrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)
type GrpcClient struct {
	Ctx context.Context
	Use bool
	Server string
	Client pb.RMSClient
	UseCodeMPAttribute bool
	Conn *grpc.ClientConn
}
type PacketEror struct{ 
	CodeMpCam     string
    SessionName   string               
    Ip            string              
    LostPacket    string               
    InvalidPacket string             
    StartTime     time.Time 
    EndTime       time.Time
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

func CreateLossСatcherClient(ctx context.Context, cfg conf.LossСatcher) (GrpcClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.GrpcAddress, cfg.GrpcPort), grpc.WithInsecure())
	if err != nil {
		return GrpcClient{}, err
	}
	client := GrpcClient{
		Ctx:    ctx,
		Use:    cfg.Use,
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

func (c *GrpcClient) Packet (paceError PacketEror) (error){

	switch {
	case !paceError.EndTime.IsZero():
		_, offset := paceError.StartTime.Zone()
		message := &pb.PacketError{
		Server:      c.Server,
		CodeMpCam:   paceError.CodeMpCam,
		Ip:          paceError.Ip,
		SessionName: paceError.SessionName,
		StartTime:   timestamppb.New(paceError.StartTime.Add(time.Duration(offset) * time.Second)),
		EndTime:     timestamppb.New(paceError.EndTime.Add(time.Duration(offset) * time.Second)),
		}
		_, err := c.Client.Packet(c.Ctx, message)
		if err != nil {
			if c.Ctx.Err() == context.Canceled {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_, err = c.Client.Packet(ctx, message)
			if err != nil {
				return err
			}
			return nil
		}
		return err
}
	case paceError.LostPacket != "":
		_, offset := paceError.StartTime.Zone()
		message:= &pb.PacketError{
			Server: c.Server,
			CodeMpCam: paceError.CodeMpCam,
			Ip: paceError.Ip,
			SessionName: paceError.SessionName,
			StartTime: timestamppb.New(paceError.StartTime.Add(time.Duration(offset) * time.Second)),
			LostPacket: paceError.LostPacket,
		}
		_, err := c.Client.Packet(c.Ctx, message)
		if err != nil {
			if c.Ctx.Err() == context.Canceled {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_, err = c.Client.Packet(ctx, message)
			if err != nil {
				return err
			}
			return nil
		}
			return err
		}
	case paceError.InvalidPacket != "":
		_, offset := paceError.StartTime.Zone()
		message:= &pb.PacketError{
			Server: c.Server,
			CodeMpCam: paceError.CodeMpCam,
			Ip: paceError.Ip,
			SessionName: paceError.SessionName,
			InvalidPacket: paceError.InvalidPacket,
			StartTime: timestamppb.New(paceError.StartTime.Add(time.Duration(offset) * time.Second)),
		}
		_, err := c.Client.Packet(c.Ctx, message)
		if err != nil {
			if c.Ctx.Err() == context.Canceled {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_, err = c.Client.Packet(ctx, message)
			if err != nil {
				return err
			}
			return nil
		}
			return err
		}
	default:
		_, offset := paceError.StartTime.Zone()
		message:= &pb.PacketError{
			Server: c.Server,
			CodeMpCam: paceError.CodeMpCam,
			Ip: paceError.Ip,
			SessionName: paceError.SessionName,
			StartTime: timestamppb.New(paceError.StartTime.Add(time.Duration(offset) * time.Second)),	
		}
		_, err := c.Client.Packet(c.Ctx, message)
		if err != nil {
			if c.Ctx.Err() == context.Canceled {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				_, err = c.Client.Packet(ctx, message)
				if err != nil {
					return err
				}
				return nil
			}
			return err
		}

	}


return nil
}