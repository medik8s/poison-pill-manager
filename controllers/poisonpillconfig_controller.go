/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"io/ioutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configpoisonpilliov1alpha1 "github.com/poison-pill/poison-pill-manager/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
)

// PoisonPillConfigReconciler reconciles a PoisonPillConfig object
type PoisonPillConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=config.poison-pill.io.poison-pill.io,resources=poisonpillconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.poison-pill.io.poison-pill.io,resources=poisonpillconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=daemonsets,verbs=get;update;patch;create
// +kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=get;update;patch;create

func (r *PoisonPillConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("poisonpillconfig", req.NamespacedName)

	content, err := ioutil.ReadFile("install/poison-pill-deamonset.yaml")
	if err != nil {
		r.Log.Error(err, "failed to read ds yaml")
		return ctrl.Result{}, err
	}
	r.Log.Info(string(content))

	decode := scheme.Codecs.UniversalDeserializer().Decode
	ppillDsObj, _, err := decode(content, nil, nil)

	if err != nil {
		r.Log.Error(err, "failed to decode poison pill daemonset yaml")
		return ctrl.Result{}, err
	}

	err = r.Client.Create(context.Background(), ppillDsObj)
	if err != nil {
		r.Log.Error(err, "failed to create poison pill daemonset")
		return ctrl.Result{}, err
	}

	r.Log.Info("reconciled succesfully")
	return ctrl.Result{}, nil
}

func (r *PoisonPillConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configpoisonpilliov1alpha1.PoisonPillConfig{}).
		Complete(r)
}
