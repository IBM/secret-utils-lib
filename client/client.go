package client

import (
	"fmt"
	"os"

	auth "github.com/IBM/secret-utils-lib/pkg/authenticator"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger := setUpLogger(false)
	authenticator, auth_type, err := auth.NewAuthenticator(logger)
	if err != nil {
		logger.Error("Error initializing authenticator")
		return
	}
	fmt.Println(auth_type)
	// To get the associated secret - (apikey/trusted-profile)
	fmt.Println(authenticator.GetSecret())
	// To set a different secret
	authenticator.SetSecret("")
	// To get a fresh token and token lifetime
	authenticator.GetToken(true)
	// To get existing token and toke lifetime
	authenticator.GetToken(false)
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
