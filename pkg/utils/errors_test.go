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

// Package utils ...
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	testcases := []struct {
		testcasename  string
		description   string
		backenderror  string
		action        string
		expectedError string
	}{
		{
			testcasename:  "All parameters populated",
			description:   "description",
			backenderror:  "backenderror",
			action:        "action",
			expectedError: "Description: description BackendError: backenderror Action: action ",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.testcasename, func(t *testing.T) {
			err := Error{
				Description:  testcase.description,
				BackendError: testcase.backenderror,
				Action:       testcase.action,
			}
			assert.Equal(t, testcase.expectedError, err.Error())
		})
	}

}
