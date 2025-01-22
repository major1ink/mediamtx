package repgrpc

import (
	"context"
	"testing"

	"github.com/bluenviron/mediamtx/internal/conf"
)



func TestCreateGrpcClient(t *testing.T) {
    ctx := context.Background()
    cfg := conf.GRPC{
        GrpcAddress: "localhost",
        GrpcPort:    50051,
        ServerName:  "test_server",
        Use:         true,
        UseCodeMPAttribute: true,
    }

    t.Run("success", func(t *testing.T) {
        client, err := CreateGrpcClient(ctx, cfg)
        if err != nil {
            t.Errorf("expected no error, got %v", err)
        }
        if client == (GrpcClient{}) {
            t.Errorf("expected non-empty client, got empty client")
        }
        if client.Ctx != ctx {
            t.Errorf("expected ctx to be %v, got %v", ctx, client.Ctx)
        }
        if client.Use != cfg.Use {
            t.Errorf("expected Use to be %v, got %v", cfg.Use, client.Use)
        }
        if client.UseCodeMPAttribute != cfg.UseCodeMPAttribute {
            t.Errorf("expected UseCodeMPAttribute to be %v, got %v", cfg.UseCodeMPAttribute, client.UseCodeMPAttribute)
        }
        if client.Server != cfg.ServerName {
            t.Errorf("expected Server to be %v, got %v", cfg.ServerName, client.Server)
        }
        if client.Client == nil {
            t.Errorf("expected non-nil Client, got nil")
        }
        if client.Conn == nil {
            t.Errorf("expected non-nil Conn, got nil")
        }
    })
}