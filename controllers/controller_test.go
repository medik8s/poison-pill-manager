package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/poison-pill/poison-pill-manager/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("Controller Test", func() {
	namespace := "default"
	dsName := "poison-pill-ds"

	config := &v1alpha1.PoisonPillConfig{}
	config.Kind = "PoisonPillConfig"
	config.APIVersion = "config.poison-pill.io.poison-pill.io/v1alpha1"
	config.Spec.WatchdogFilePath = "/dev/foo"
	config.Name = "config-sample"
	config.Namespace = namespace

	It("Config CR should be created", func() {
		Expect(k8sClient).To(Not(BeNil()))
		Expect(k8sClient.Create(context.Background(), config)).To(Succeed())
		createdConfig := &v1alpha1.PoisonPillConfig{}
		configKey, err := client.ObjectKeyFromObject(config)
		Expect(err).To(BeNil())
		Expect(k8sClient.Get(context.Background(), configKey, createdConfig)).To(Succeed())
		Expect(createdConfig.Spec.WatchdogFilePath).To(Equal(config.Spec.WatchdogFilePath))
		Expect(createdConfig.Spec.SafeTimeToAssumeNodeRebootedSeconds).To(Equal(config.Spec.SafeTimeToAssumeNodeRebootedSeconds))
	})

	It("Daemonset should be created", func() {
		ds := &v1.DaemonSet{}
		key := types.NamespacedName{
			Namespace: namespace,
			Name:      dsName,
		}
		Eventually(func() error {
			return k8sClient.Get(context.Background(), key, ds)
		}, 5*time.Second, 250*time.Millisecond).Should(BeNil())

	})

})
