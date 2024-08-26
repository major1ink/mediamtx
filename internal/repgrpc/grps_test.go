package repgrpc

import (
	"context"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/golang/mock/gomock"
	pb "github.com/major1ink/repGrpc/pkg/repGrpc"
	"google.golang.org/protobuf/types/known/emptypb"
)


func TestSelect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockGrpcClient(ctrl)
	c := &GrpcClient{
		Client: mockClient,
		Server: "test_server",
		Ctx:    context.Background(),
	}

	testSelect := []struct {
		streamName string
		argument   string
		expected   *pb.AnswerSelect
		err        error
	}{
		{
			streamName: "stream1",
			argument:   "CodeMP",
			expected:   &pb.AnswerSelect{CodeMP: "123" , 
		StatusRecord: 1,},
			err:        nil,
		},
		{
			streamName: "stream2",
			argument:   "CodeMP_Contract",
			expected:   &pb.AnswerSelect{ CodeMPContract: "123" },
			err:        nil,
		},
		{
			streamName: "stream3",
			argument:   "MountPoint",
			expected:   &pb.AnswerSelect{ MapDisks: map[string]int32{"./recordings": 2, "./recordings2": 2} },
			err:        nil,
		},
		{
			streamName: "stream4",
			argument:   "StatusRecord",
			expected:   &pb.AnswerSelect{ StatusRecord: 1 },
			err:        nil,
		},
	}

	for _, tc := range testSelect {
		t.Run(tc.argument, func(t *testing.T) {
			mockClient.EXPECT().Get(c.Ctx, gomock.Any()).Return(tc.expected, tc.err)

			result, err := c.Select(tc.streamName, tc.argument)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, result)
		})
	}
	testInsert := []struct {
		attribute   string
		query string
		expected   *emptypb.Empty
		err        error
	}{
	{
			attribute:   "PathStream",
			query: "(\"1\", \"2\", \"3\")",
			expected:   &emptypb.Empty{},
			err:        nil,
	},
	{
			attribute:   "Stream",
			query: "(\"1\", \"2\", \"3\")",
			expected:   &emptypb.Empty{},
			err:        nil,
	},
	{
			attribute:   "CodeMP",
			query: "(\"1\", \"2\", \"3\")",
			expected:   &emptypb.Empty{},
			err:        nil,
	},
}

for _, tc := range testInsert {
	t.Run(tc.attribute, func(t *testing.T) {
			mockClient.EXPECT().Post(c.Ctx, gomock.Any()).Return(tc.expected, tc.err)

			 err := c.Post(tc.attribute, tc.query)
			assert.Equal(t, tc.err, err)
	})
}
}

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