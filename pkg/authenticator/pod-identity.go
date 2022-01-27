package authenticator

import (
	"IBM/secret-utils-lib/pkg/token"
	"IBM/secret-utils-lib/pkg/utils"

	//"github.com/GunaKKIBM/secret-utils-lib/pkg/token"
	//"github.com/GunaKKIBM/secret-utils-lib/pkg/utils"
	"github.com/IBM/go-sdk-core/v5/core"
	"go.uber.org/zap"
)

// ComputeIdentityAuthenticator ...
type ComputeIdentityAuthenticator struct {
	authenticator *core.ContainerAuthenticator
	logger        *zap.Logger
}

// NewComputeIdentityAuthenticator ...
func NewComputeIdentityAuthenticator(profileID string, logger *zap.Logger) *ComputeIdentityAuthenticator {
	logger.Info("Initializing compute identity authenticator")
	defer logger.Info("Initialized compute identity authenticator")
	ca := new(ComputeIdentityAuthenticator)
	ca.authenticator = new(core.ContainerAuthenticator)
	ca.authenticator.IAMProfileID = profileID
	ca.logger = logger
	return ca
}

// GetToken ...
func (ca *ComputeIdentityAuthenticator) GetToken(freshTokenRequired bool) (string, uint64, error) {
	ca.logger.Info("Fetching token using compute identity authenticator")
	var iamtoken string
	var err error
	var tokenlifetime uint64

	if !freshTokenRequired {
		ca.logger.Info("Retreiving existing token")
		iamtoken, err = ca.authenticator.GetToken()
		if err != nil {
			ca.logger.Error("Error fetching iam token", zap.Error(err))
			return "", tokenlifetime, err
		}
		tokenlifetime, err = token.FetchTokenLifeTime(iamtoken)
		if err != nil {
			ca.logger.Error("Error fetching token lifetime", zap.Error(err))
			return "", tokenlifetime, err
		}
		if tokenlifetime > utils.TokenExpirydiff {
			ca.logger.Info("Successfully fetched iam token")
			return iamtoken, tokenlifetime, nil
		}
	}

	tokenResponse, err := ca.authenticator.RequestToken()
	if err != nil {
		ca.logger.Error("Error fetching fresh iam token", zap.Error(err))
		return "", tokenlifetime, err
	}

	tokenlifetime, err = token.FetchTokenLifeTime(tokenResponse.AccessToken)
	if err != nil {
		ca.logger.Error("Error fetching token lifetime", zap.Error(err))
		return "", tokenlifetime, err
	}
	ca.logger.Info("Successfully fetched iam token")
	return tokenResponse.AccessToken, tokenlifetime, nil
}

// GetSecret ...
func (ca *ComputeIdentityAuthenticator) GetSecret() string {
        return ca.authenticator.IAMProfileID
}

// SetSecret ...
func (aa *ComputeIdentityAuthenticator) SetSecret(secret string) {
        ca.authenticator.IAMProfileID = secret
}
