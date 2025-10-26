package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	pb_account "github.com/PretendoNetwork/grpc/go/account"
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/plogger-go"
	"github.com/hoverAdev/xenoblade-chronicles-x/globals"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func init() {
	globals.Logger = plogger.NewLogger()

	var err error

	err = godotenv.Load()
	if err != nil {
		globals.Logger.Warning("Error loading .env file")

		s3Endpoint := os.Getenv("PN_XCX_CONFIG_S3_ENDPOINT")
		s3AccessKey := os.Getenv("PN_XCX_CONFIG_S3_ACCESS_KEY")
		s3AccessSecret := os.Getenv("PN_XCX_CONFIG_S3_ACCESS_SECRET")

		authenticationServerPort := os.Getenv("PN_XCX_AUTHENTICATION_SERVER_PORT")
		secureServerHost := os.Getenv("PN_XCX_SECURE_SERVER_HOST")
		secureServerPort := os.Getenv("PN_XCX_SECURE_SERVER_PORT")
		accountGRPCHost := os.Getenv("PN_XCX_ACCOUNT_GRPC_HOST")
		accountGRPCPort := os.Getenv("PN_XCX_ACCOUNT_GRPC_PORT")
		accountGRPCAPIKey := os.Getenv("PN_XCX_ACCOUNT_GRPC_API_KEY")
		tokenAesKey := os.Getenv("PN_XCX_AES_KEY")
		localAuthMode := os.Getenv("PN_XCX_LOCAL_AUTH")

		kerberosPassword := make([]byte, 0x10)
		_, err = rand.Read(kerberosPassword)
		if err != nil {
			globals.Logger.Error("Error generating Kerberos password")
			os.Exit(0)
		}

		globals.KerberosPassword = string(kerberosPassword)

		globals.AuthenticationServerAccount = nex.NewAccount(types.NewPID(1), "Quazal Authentication", globals.KerberosPassword, false)
		globals.SecureServerAccount = nex.NewAccount(types.NewPID(2), "Quazal Rendez-Vous", globals.KerberosPassword, false)

		if strings.TrimSpace(authenticationServerPort) == "" {
			globals.Logger.Error("PN_XCX_AUTHENTICATION_SERVER_PORT environment variable not set")
			os.Exit(0)
		}

		if port, err := strconv.Atoi(authenticationServerPort); err != nil {
			globals.Logger.Errorf("PN_XCX_AUTHENTICATION_SERVER_PORT is not a valid port. Expected 0-65535, got %s", authenticationServerPort)
			os.Exit(0)
		} else if port < 0 || port > 65535 {
			globals.Logger.Errorf("PN_XCX_AUTHENTICATION_SERVER_PORT is not a valid port. Expected 0-65535, got %d", port)
		}

		if strings.TrimSpace(secureServerHost) == "" {
			globals.Logger.Error("PN_XCX_SECURE_SERVER_HOST environment variable not set")
			os.Exit(0)
		}

		if strings.TrimSpace(secureServerPort) == "" {
			globals.Logger.Error("PN_XCX_SECURE_SERVER_PORT environment variable not set")
			os.Exit(0)
		}

		if port, err := strconv.Atoi(secureServerPort); err != nil {
			globals.Logger.Errorf("PN_XCX_SECURE_SERVER_PORT is not a valid port. Expected 0-65535, got %s", secureServerPort)
			os.Exit(0)
		} else if port < 0 || port > 65535 {
			globals.Logger.Errorf("PN_XCX_SECURE_SERVER_PORT is not a valid port. Expected 0-65535, got %d", secureServerPort)
			os.Exit(0)
		}

		if strings.TrimSpace(accountGRPCHost) == "" {
			globals.Logger.Error("PN_XCX_ACCOUNT_GRPC_HOST environment variable not set")
			os.Exit(0)
		}

		if strings.TrimSpace(accountGRPCPort) == "" {
			globals.Logger.Error("PN_XCX_ACCOUNT_GRPC_PORT environment variable not set")
			os.Exit(0)
		}

		if port, err := strconv.Atoi(accountGRPCPort); err != nil {
			globals.Logger.Errorf("PN_XCX_ACCOUNT_GRPC_PORT is not a valid port. Expected 0-65535, got %s", accountGRPCPort)
			os.Exit(0)
		} else if port < 0 || port > 65535 {
			globals.Logger.Errorf("PN_XCX_ACCOUNT_GRPC_PORT is not a valid port. Expected 0-65535, got %d", accountGRPCPort)
			os.Exit(0)
		}

		if strings.TrimSpace(accountGRPCAPIKey) == "" {
			globals.Logger.Warning("Insecure gRPC server detected. PN_XCX_ACCOUNT_GRPC_API_KEY environment variable not set")
		}

		globals.GRPCAccountClientConnection, err = grpc.NewClient(fmt.Sprintf("%s:%s", accountGRPCHost, accountGRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			globals.Logger.Criticalf("Failed to connect to account gRPC server: %v", err)
			os.Exit(0)
		}

		globals.GRPCAccountClient = pb_account.NewAccountClient(globals.GRPCAccountClientConnection)
		globals.GRPCAccountCommonMetadata = metadata.Pairs(
			"X-API-Key", accountGRPCAPIKey,
		)

		staticCredentials := credentials.NewStaticV4(s3AccessKey, s3AccessSecret, "")

		minIOClient, err := minio.New(s3Endpoint, &minio.Options{
			Creds:  staticCredentials,
			Secure: true,
		})
		if err != nil {
			globals.Logger.Criticalf("Error occured during minio connection: %v", err)
			os.Exit(0)
		}

		globals.MinIOClient = minIOClient
		globals.Presigner = globals.NewS3Presigner(globals.MinIOClient)

		if strings.TrimSpace(tokenAesKey) == "" {
			globals.Logger.Error("PN_XC_AES_KEY environment variable not set")
			os.Exit(0)
		}

		globals.TokenAESKey, err = hex.DecodeString(tokenAesKey)
		if err != nil {
			globals.Logger.Errorf("Failed to decode AES key: %v", err)
			os.Exit(0)
		}

		globals.LocalAuthMode = localAuthMode == "1"
		if globals.LocalAuthMode {
			globals.Logger.Warning("Local authentication mode is enabled. Token validation will be skipped!")
			globals.Logger.Warning("This is insecure and could allow ban bypasses!")
		}
	}
}
