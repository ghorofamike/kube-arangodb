//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package reconcile

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type gracefulAction struct {
	actionEmpty

	graceful time.Duration
}

func (g gracefulAction) StartFailureGracePeriod() time.Duration {
	return g.graceful
}

var _ ActionStartFailureGracePeriod = gracefulAction{}

func Test_GracefulTimeouts(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		require.EqualValues(t, 0, getStartFailureGracePeriod(actionEmpty{}))
	})
	t.Run("Set", func(t *testing.T) {
		require.EqualValues(t, time.Second, getStartFailureGracePeriod(gracefulAction{
			graceful: time.Second,
		}))
	})
	t.Run("Override", func(t *testing.T) {
		require.EqualValues(t, time.Minute, getStartFailureGracePeriod(wrapActionStartFailureGracePeriod(gracefulAction{
			graceful: time.Second,
		}, time.Minute)))
	})
}
