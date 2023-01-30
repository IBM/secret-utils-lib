# secret-utils-lib
This library is used for fetching iam token using two possible methods
- Using Trusted profile
- Using API key

The client code in client directory shows how this library can be used.

## Pre requisites

- A k8s secret must be present in the same namespace where the pod (the application in which this code is used) is deployed.
- The secrets format are present in secrets folder - ibm-cloud-credentials.yaml or storage-secret-store.yaml.
- To use the trusted profile, the following needs to be added in the deployment file of the application that is using this library.
  1. Under the volumes, the following entity needs to be added
    ```
    volumes:
    - name: vault-token
        projected:
        sources:
        - serviceAccountToken:
            path: vault-token
            expirationSeconds: 600
    ```
  2. The following volume mount needs to be added to the container application which is using this library
    ```
    volumeMounts:
    - mountPath: /var/run/secrets/tokens
      name: vault-token
    ```

## Functionality
The library code first looks for [ibm-cloud-credentials](https://github.com/IBM/secret-utils-lib/tree/master/secrets/ibm-cloud-credentials) k8s secret and reads trusted-profile-id/api-key from it. If ibm-cloud-credentials is not present, the code looks for [storage-secret-store](https://github.com/IBM/secret-utils-lib/tree/master/secrets/storage-secret-store) k8s secret and reads api-key from the same. Later when the required method is called to fetch the iam token the same trusted-profile/api-key is used for fetching the token. More details are shared below.

## Methods defined to initialize and use the authenticators

**Note:** This library is designed so that it can be used by [secret-common-lib](https://github.com/IBM/secret-common-lib). It is recommended to use the same instead for directly using this library.

### Initializing the authenticator

```
NewAuthenticator(logger *zap.Logger, kc k8s_utils.KubernetesClient, optionalArgs ...map[string]string) (Authenticator, string, error) (Authenticator, authType, error)
```

As seen above, authenticator can be initialized using the `NewAuthenticator` method which needs three mandatory arguments and one optional argument.
- `logger`: Pass an initialized [zap.Logger](https://pkg.go.dev/go.uber.org/zap#Logger) object.
- `KubernetesClient`: Pass an initialized [kubernetes client object](https://github.com/IBM/secret-utils-lib/blob/master/pkg/k8s_utils/k8s_client.go#L52).
- `providerName`: This needs to be either one of `vpc`, `bluemix`, `softlayer`. It is needed because the library needs to know where to read the api-key from in case of using [storage-secret-store](https://github.com/IBM/secret-utils-lib/blob/master/secrets/storage-secret-store/slclient.toml)
- `optionalArgs`: This is an optional argument. If the client is using storage-secret-store, the argument here should look like map[ProviderType]value, where value should be either `vpc`, `bluemix`, `softlayer` OR If the client using this library doesn't want to use the default keys in secret(which is [ibm-credentials.env](https://github.com/IBM/secret-utils-lib/blob/master/secrets/ibm-cloud-credentials/ibm-cloud-credentials.yaml#L3) in ibm-cloud-credentials and [slclient.toml](https://github.com/IBM/secret-utils-lib/blob/master/secrets/storage-secret-store/storage-secret-store.yaml#L3) in storage-secret-store), there is another option of having specific keys in either ibm-cloud-credentials or storage-secret-store.
- If specific key is provided in ibm-cloud-credentials, it must be provided as base64 encoded value of [this](https://github.com/IBM/secret-utils-lib/blob/master/secrets/ibm-cloud-credentials/iam-cloud-provider.env) format itself and the k8s secret looks like [this](https://github.com/IBM/secret-utils-lib/blob/master/secrets/ibm-cloud-credentials/ibm-cloud-credentials-with-secret-key.yaml).
- If specific key is provided in storage-secret-store, it must be provided as base64 encoded value of api-key and the k8s secret looks like [this](https://github.com/IBM/secret-utils-lib/blob/master/secrets/storage-secret-store/storage-secret-store-with-key.yaml).
- **Note:** The library first looks for the `secretKey` in `ibm-cloud-credentials`, if it doesn't exist there, it is searched in `storage-secret-store`. So, if the application using this library has a use case of using `secretKey`, we recommend to name them differently for ibm-cloud-credentials and storage-secret-store.
- The client functions [here](https://github.com/IBM/secret-utils-lib/blob/master/client/client.go) show how the authenticator can be initialised and used.

### Fetching the token.

IAM token for the trusted-profile-id/api-key can be fetched by calling the `GetToken` method with reference to the initialized authenticator. Please refer the [client code examples](https://github.com/IBM/secret-utils-lib/blob/master/client/client.go).

Methods supported by the authenticator.
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


