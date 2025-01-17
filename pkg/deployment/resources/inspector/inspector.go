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

package inspector

import (
	"context"
	"sync"

	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretReadInterface has methods to work with Secret resources with ReadOnly mode.
type SecretReadInterface interface {
	Get(ctx context.Context, name string, opts meta.GetOptions) (*core.Secret, error)
}

func NewInspector(ctx context.Context, client kclient.Client, namespace string) (inspectorInterface.Inspector, error) {
	i := &inspector{
		namespace: namespace,
		client:    client,
	}

	if err := i.Refresh(ctx); err != nil {
		return nil, err
	}

	return i, nil
}

func newInspector(ctx context.Context, client kclient.Client, namespace string) (*inspector, error) {
	var i inspector

	i.namespace = namespace
	i.client = client

	if err := util.RunParallel(15,
		getVersionInfo(ctx, &i, client.Kubernetes(), namespace),
		podsToMap(ctx, &i, client.Kubernetes(), namespace),
		secretsToMap(ctx, &i, client.Kubernetes(), namespace),
		pvcsToMap(ctx, &i, client.Kubernetes(), namespace),
		servicesToMap(ctx, &i, client.Kubernetes(), namespace),
		serviceAccountsToMap(ctx, &i, client.Kubernetes(), namespace),
		podDisruptionBudgetsToMap(ctx, &i, client.Kubernetes(), namespace),
		serviceMonitorsToMap(ctx, &i, client.Monitoring(), namespace),
		arangoMembersToMap(ctx, &i, client.Arango(), namespace),
		nodesToMap(ctx, &i, client.Kubernetes()),
		arangoClusterSynchronizationsToMap(ctx, &i, client.Arango(), namespace),
		arangoTasksToMap(ctx, &i, client.Arango(), namespace),
	); err != nil {
		return nil, err
	}

	return &i, nil
}

func NewEmptyInspector() inspectorInterface.Inspector {
	return NewInspectorFromData(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "")
}

func NewInspectorFromData(pods map[string]*core.Pod,
	secrets map[string]*core.Secret,
	pvcs map[string]*core.PersistentVolumeClaim,
	services map[string]*core.Service,
	serviceAccounts map[string]*core.ServiceAccount,
	podDisruptionBudgets map[string]*policy.PodDisruptionBudget,
	serviceMonitors map[string]*monitoring.ServiceMonitor,
	arangoMembers map[string]*api.ArangoMember,
	nodes map[string]*core.Node,
	acs map[string]*api.ArangoClusterSynchronization,
	at map[string]*api.ArangoTask,
	version driver.Version) inspectorInterface.Inspector {
	i := &inspector{
		pods:                 pods,
		secrets:              secrets,
		pvcs:                 pvcs,
		services:             services,
		serviceAccounts:      serviceAccounts,
		podDisruptionBudgets: podDisruptionBudgets,
		serviceMonitors:      serviceMonitors,
		arangoMembers:        arangoMembers,
		versionInfo:          version,
	}

	if nodes == nil {
		i.nodes = &nodeLoader{
			accessible: false,
			nodes:      nil,
		}
	} else {
		i.nodes = &nodeLoader{
			accessible: true,
			nodes:      nodes,
		}
	}

	if acs == nil {
		i.acs = &arangoClusterSynchronizationLoader{
			accessible: false,
			acs:        nil,
		}
	} else {
		i.acs = &arangoClusterSynchronizationLoader{
			accessible: true,
			acs:        acs,
		}
	}

	if at == nil {
		i.at = &arangoTaskLoader{
			accessible: false,
			at:         nil,
		}
	} else {
		i.at = &arangoTaskLoader{
			accessible: true,
			at:         at,
		}
	}

	return i
}

type inspector struct {
	lock sync.Mutex

	namespace string

	client kclient.Client

	pods                 map[string]*core.Pod
	secrets              map[string]*core.Secret
	pvcs                 map[string]*core.PersistentVolumeClaim
	services             map[string]*core.Service
	serviceAccounts      map[string]*core.ServiceAccount
	podDisruptionBudgets map[string]*policy.PodDisruptionBudget
	serviceMonitors      map[string]*monitoring.ServiceMonitor
	arangoMembers        map[string]*api.ArangoMember
	nodes                *nodeLoader
	acs                  *arangoClusterSynchronizationLoader
	at                   *arangoTaskLoader
	versionInfo          driver.Version
}

func (i *inspector) IsStatic() bool {
	return i.namespace == ""
}

func (i *inspector) Refresh(ctx context.Context) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.namespace == "" {
		return errors.New("Inspector created from static data")
	}

	new, err := newInspector(ctx, i.client, i.namespace)
	if err != nil {
		return err
	}

	i.pods = new.pods
	i.secrets = new.secrets
	i.pvcs = new.pvcs
	i.services = new.services
	i.serviceAccounts = new.serviceAccounts
	i.podDisruptionBudgets = new.podDisruptionBudgets
	i.serviceMonitors = new.serviceMonitors
	i.arangoMembers = new.arangoMembers
	i.nodes = new.nodes
	i.acs = new.acs
	i.versionInfo = new.versionInfo

	return nil
}
