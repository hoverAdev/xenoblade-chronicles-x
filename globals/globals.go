package globals

import (
	pb_account "github.com/PretendoNetwork/grpc/go/account"
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/plogger-go"
	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var Logger *plogger.Logger
var KerberosPassword = "password"

var AuthenticationServer *nex.PRUDPServer
var AuthenticationEndpoint *nex.PRUDPEndPoint

var SecureServer *nex.PRUDPServer
var SecureEndpoint *nex.PRUDPEndPoint

var GRPCAccountClientConnection *grpc.ClientConn
var GRPCAccountClient pb_account.AccountClient
var GRPCAccountCommonMetadata metadata.MD

var MinIOClient *minio.Client
var Presigner *S3Presigner

var TokenAESKey []byte
var LocalAuthMode bool
