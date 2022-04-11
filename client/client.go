package main

import (
	"fmt"
	"os"
	"path/filepath"

	auth "github.com/IBM/secret-utils-lib/pkg/authenticator"
	"github.com/IBM/secret-utils-lib/pkg/k8s_utils"
	"github.com/IBM/secret-utils-lib/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger := setUpLogger(false)
	client, _ := k8s_utils.FakeGetk8sClientSet(logger)
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error fetching current working directory")
		return
	}
	// secretFilePath path to the content where the secret data is stored

	/*secretFilePath := filepath.Join(cwd, "..", "secrets/storage-secret-store/slclient.toml")
	err = k8s_utils.FakeCreateSecret(client, utils.DEFAULT, secretFilePath)*/

	secretFilePath := filepath.Join(cwd, "..", "secrets/ibm-cloud-credentials/iam-cloud-provider.env")
	err = k8s_utils.FakeCreateSecret(client, utils.IAM, secretFilePath)
	if err != nil {
		logger.Error("Error creating secret", zap.Error(err))
		return
	}

	authenticator, auth_type, err := auth.NewAuthenticator(logger, client)
	if err != nil {
		logger.Error("Error initializing authenticator")
		return
	}

	fmt.Println(auth_type)
	// To get the associated secret - (apikey/trusted-profile)
	fmt.Println(authenticator.GetSecret())
	// To set a different secret
	//authenticator.SetSecret("")
	// To get token and token lifetime
	fmt.Println(authenticator.GetToken(false))
}

// setUpLogger ...
func setUpLogger(managed bool) *zap.Logger {
	// Prepare a new logger
	atom := zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	var secretProviderType string
	if managed {
		secretProviderType = "managed-secret-provider"
	} else {
		secretProviderType = "unmanaged-secret-provider"
	}
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	), zap.AddCaller()).With(zap.String("name", "secret-provider")).With(zap.String("secret-provider-type", secretProviderType))

	atom.SetLevel(zap.InfoLevel)
	return logger
}
