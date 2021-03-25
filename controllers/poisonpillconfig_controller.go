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
	"k8s.io/apimachinery/pkg/api/errors"
	"log"
	"path"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configpoisonpilliov1alpha1 "github.com/poison-pill/poison-pill-manager/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
)

// PoisonPillConfigReconciler reconciles a PoisonPillConfig object
type PoisonPillConfigReconciler struct {
	installFileFolder string
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=config.medik8s.io,resources=poisonpillconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.medik8s.io,resources=poisonpillconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=machine.openshift.io,resources=machines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=machine.openshift.io,resources=machines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;delete;create
// +kubebuilder:rbac:groups="",resources=daemonsets,verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups="extensions",resources=daemonsets,verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;update;patch;create;escalate
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;update;patch;create;escalate
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles,verbs=get;update;patch;create;escalate
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;update;patch;create;escalate
// +kubebuilder:rbac:groups="",resources=services,verbs=get;update;patch;create
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;update;patch;create
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;update;patch;create
// +kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,verbs=use,resourceNames=privileged
// +kubebuilder:rbac:groups="config.openshift.io",resources=infrastructures,verbs=get;list;watch

func (r *PoisonPillConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("poisonpillconfig", req.NamespacedName)

	filePath := path.Join(r.installFileFolder, "install/poison-pill-deamonset-with-rbac.yaml")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		r.Log.Error(err, "failed to read ds yaml")
		return ctrl.Result{}, err
	}
	r.Log.Info(string(content))

	objects := r.parseK8sYaml(content)
	for _, obj := range objects {
		err = r.Client.Create(context.Background(), obj)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				err = r.Client.Update(context.Background(), obj)
				if err != nil {
					r.Log.Error(err, "failed to update poison pill runtime object", "kind", obj.GetObjectKind().GroupVersionKind().Kind)
					return ctrl.Result{}, err
				}
			} else {
				r.Log.Error(err, "failed to create poison pill runtime object", "kind", obj.GetObjectKind().GroupVersionKind().Kind)
				return ctrl.Result{}, err
			}
		}
	}

	r.Log.Info("reconciled successfully")
	return ctrl.Result{}, nil
}

func (r *PoisonPillConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configpoisonpilliov1alpha1.PoisonPillConfig{}).
		Complete(r)
}

func (r *PoisonPillConfigReconciler) parseK8sYaml(fileR []byte) []runtime.Object {
	acceptedK8sTypes := regexp.MustCompile(`(Namespace|Role|ClusterRole|RoleBinding|ClusterRoleBinding|ServiceAccount|DaemonSet|Service)`)
	fileAsString := string(fileR[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	retVal := make([]runtime.Object, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(f), nil, nil)

		if err != nil {
			r.Log.Error(err, "error while decoding YAML object")
			continue
		}

		if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
			log.Printf("the custom-roles configMap contained object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind)
		} else {
			retVal = append(retVal, obj)
		}

	}
	return retVal
}
