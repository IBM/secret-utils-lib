# secret-utils-lib
This library is used for fetching iam token using two possible methods
- Using Compute Identity
- Using API key

The client code in client directory shows how this library can be used

### Methods defined for authenticators

```
// GetToken returns iam token, token lifetime and error if any
// if freshTokenRequired is set to true, a call is made to iam to fetch a fresh token and returned
// else, the token stored in cache is validated, if valid, the same is returned (hence avoiding the call to iam), else a call is made to iam to fetch a fresh token
GetToken(freshTokenRequired bool) (string, uint64, error)

// GetSecret returns the appropriate secret based on the type of authenticator
GetSecret() string

// SetSecret modifies the existing secret (removes existing secret and sets the new secret)
SetSecret(secret string)
```

### Pre requisites

- A k8s secret must be present in the same namespace where the pod (the application in which this code is used) is deployed.
- The secrets format are present in secrets folder - ibm-cloud-credentials.yaml or storage-secret-store.yaml
