/**
 * Copyright 2021 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package authenticator ...
package authenticator

import "go.uber.org/zap"

type FakeAuthenticator struct {
	secret string
	logger *zap.Logger
}

func newFakeAuthenticator(secret string, logger *zap.Logger) *FakeAuthenticator {
	return &FakeAuthenticator{secret: secret, logger: logger}
}

func (fa *FakeAuthenticator) GetToken() (string, uint64, error) {
	return "", 0, nil
}

func (fa *FakeAuthenticator) GetSecret() string {
	return fa.secret
}

func (fa *FakeAuthenticator) SetSecret(secret string) {
	fa.secret = secret
}
