package conf

type GRPC struct {
	Use bool `json:"use"`
	GrpcAddress string `json:"grpcAddress"`
	GrpcPort int `json:"grpcPort"`
	ServerName string `json:"serverName"`
}

func (grpc *GRPC) setDefaults() {
	grpc.Use = false
	grpc.GrpcAddress = "127.0.0.1"
	grpc.GrpcPort = 8080
	grpc.ServerName = "RMS"
}