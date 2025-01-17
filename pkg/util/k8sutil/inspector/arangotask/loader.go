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

package arangotask

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type Loader interface {
	GetArangoTasks() (Inspector, bool)
}

type Inspector interface {
	ArangoTasks() []*api.ArangoTask
	ArangoTask(name string) (*api.ArangoTask, bool)
	FilterArangoTasks(filters ...Filter) []*api.ArangoTask
	IterateArangoTasks(action Action, filters ...Filter) error
	ArangoTaskReadInterface() ReadInterface
}

type Filter func(acs *api.ArangoTask) bool
type Action func(acs *api.ArangoTask) error
