/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2022 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// Package secret_provider ...
package secret_provider

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFakeGetDefaultIAMToken ...
func TestFakeGetDefaultIAMToken(t *testing.T) {
	testcases := []struct {
		testcasename         string
		isFreshTokenRequired bool
		expectedError        error
	}{
		{
			testcasename:         "Successfully fetched token",
			isFreshTokenRequired: true,
			expectedError:        nil,
		},
		{
			testcasename:         "Error fetching token",
			isFreshTokenRequired: false,
			expectedError:        errors.New("not nil"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.testcasename, func(t *testing.T) {
			fs := new(FakeSecretProvider)
			_, _, err := fs.GetDefaultIAMToken(testcase.isFreshTokenRequired)
			if testcase.expectedError != nil {
				assert.NotNil(t, err, testcase.expectedError)
			}
		})
	}
}

// TestFakeGetIAMToken ...
func TestFakeGetIAMToken(t *testing.T) {
	testcases := []struct {
		testcasename         string
		isFreshTokenRequired bool
		expectedError        error
	}{
		{
			testcasename:         "Successfully fetched token",
			isFreshTokenRequired: true,
			expectedError:        nil,
		},
		{
			testcasename:         "Error fetching token",
			isFreshTokenRequired: false,
			expectedError:        errors.New("not nil"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.testcasename, func(t *testing.T) {
			fs := new(FakeSecretProvider)
			_, _, err := fs.GetIAMToken("fake-secret", testcase.isFreshTokenRequired)
			if testcase.expectedError != nil {
				assert.NotNil(t, err, testcase.expectedError)
			}
		})
	}
}
