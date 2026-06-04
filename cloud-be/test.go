package cloud_be

import (
	"context"
	"fmt"
)

/*
Copyright 2021 The Kubernetes Authors.

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
"fmt"

. "github.com/onsi/ginkgo/v2"
. "github.com/onsi/gomega"
corev1 "k8s.io/api/core/v1"
"k8s.io/apimachinery/pkg/api/resource"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
"k8s.io/apimachinery/pkg/types"
kubevirtv1 "kubevirt.io/api/core/v1"
ctrl "sigs.k8s.io/controller-runtime"
"sigs.k8s.io/controller-runtime/pkg/client/fake"

infrav1 "sigs.k8s.io/cluster-api-provider-kubevirt/api/v1alpha1"
testutil "sigs.k8s.io/cluster-api-provider-kubevirt/pkg/testing"
)

var _ = Describe("KubevirtMachineTemplateReconciler - extractCapacity", func() {
	reconciler := &KubevirtMachineTemplateReconciler{}
	DescribeTable("should extract correct capacity",
		func(mt infrav1.KubevirtMachineTemplate, expected corev1.ResourceList) {
			capacity, err := reconciler.extractCapacity(mt)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(capacity)).To(Equal(len(expected)))
			for key, expectedVal := range expected {
				actualVal, ok := capacity[key]
				Expect(ok).To(BeTrue(), fmt.Sprintf("expected key %s not found in capacity", key))
				Expect(expectedVal.Equal(actualVal)).To(BeTrue(), fmt.Sprintf("for key %s: expected %v, got %v", key, expectedVal, actualVal))
			}
		},

		Entry("empty template should return empty capacity", newMachineTemplate(nil, nil, nil), corev1.ResourceList{}),
		Entry("vmTemplate.Spec.Template == nil should return empty capacity",
			func() infrav1.KubevirtMachineTemplate {
				mt := newMachineTemplate(nil, nil, nil)
				mt.Spec.Template.Spec.VirtualMachineTemplate.Spec.Template = nil
				return mt
			}(),
			corev1.ResourceList{},
		),
		Entry("resources from domain.resources.requests",
			newMachineTemplate(
				&kubevirtv1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				},
				nil, nil,
			),
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2"), corev1.ResourceMemory: resource.MustParse("4Gi")},
		),

		Entry("CPU from domain.cpu.cores",
			newMachineTemplate(nil, &kubevirtv1.CPU{Cores: 4}, nil),
			corev1.ResourceList{corev1.ResourceCPU: *resource.NewQuantity(4, resource.DecimalSI)},
		),

		Entry("memory from domain.memory.guest",
			newMachineTemplate(nil, nil, &kubevirtv1.Memory{Guest: resourcePtr("8Gi")}),
			corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
		),

		Entry("cpu.cores takes precedence over requests, requests take precedence over memory.guest",
			newMachineTemplate(
				&kubevirtv1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				},
				&kubevirtv1.CPU{Cores: 8},
				&kubevirtv1.Memory{Guest: resourcePtr("16Gi")},
			),
			corev1.ResourceList{corev1.ResourceCPU: *resource.NewQuantity(8, resource.DecimalSI), corev1.ResourceMemory: resource.MustParse("4Gi")},
		),

		Entry("limits are only evaluated for cpu and memory, ignoring other resources",
			newMachineTemplate(
				&kubevirtv1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
					Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				},
				nil, nil,
			),
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2"), corev1.ResourceMemory: resource.MustParse("4Gi")},
		),

		Entry("cpu from cores when not in requests, memory from requests",
			newMachineTemplate(
				&kubevirtv1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")}},
				&kubevirtv1.CPU{Cores: 4},
				nil,
			),
			corev1.ResourceList{corev1.ResourceCPU: *resource.NewQuantity(4, resource.DecimalSI), corev1.ResourceMemory: resource.MustParse("4Gi")},
		),

		Entry("cpu multiplication from cores*sockets*threads",
			newMachineTemplate(nil, &kubevirtv1.CPU{Cores: 2, Sockets: 3, Threads: 4}, nil),
			corev1.ResourceList{corev1.ResourceCPU: *resource.NewQuantity(24, resource.DecimalSI)},
		),

		Entry("limits used as final fallback when requests and domain CPU/memory absent",
			newMachineTemplate(
				&kubevirtv1.ResourceRequirements{
					Requests: corev1.ResourceList{},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("3"),
						corev1.ResourceMemory: resource.MustParse("6Gi"),
					},
				},
				nil, nil,
			),
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("3"), corev1.ResourceMemory: resource.MustParse("6Gi")},
		),
	)
})

var _ = Describe("KubevirtMachineTemplateReconciler - Reconcile", func() {
	It("should update status.capacity based on the VM template", func() {
		// Create a machine template with CPU cores=2
		mt := newMachineTemplate(nil, &kubevirtv1.CPU{Cores: 2}, nil)

		// Build a fake client with status subresource enabled so Patch via helper works
		fakeClient := fake.NewClientBuilder().WithScheme(testutil.SetupScheme()).WithObjects(&mt).WithStatusSubresource(&mt).Build()

		reconciler := &KubevirtMachineTemplateReconciler{Client: fakeClient, Log: ctrl.Log.WithName("test")}

		req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: mt.Namespace, Name: mt.Name}}
		_, err := reconciler.Reconcile(context.TODO(), req)
		Expect(err).NotTo(HaveOccurred())

		// Read the object back and assert status.capacity was updated
		updated := &infrav1.KubevirtMachineTemplate{}
		Expect(fakeClient.Get(context.TODO(), types.NamespacedName{Namespace: mt.Namespace, Name: mt.Name}, updated)).To(Succeed())

		expected := corev1.ResourceList{corev1.ResourceCPU: *resource.NewQuantity(2, resource.DecimalSI)}
		Expect(len(updated.Status.Capacity)).To(Equal(len(expected)))
		for k, ev := range expected {
			av, ok := updated.Status.Capacity[k]
			Expect(ok).To(BeTrue())
			Expect(ev.Equal(av)).To(BeTrue(), fmt.Sprintf("expected %v for %s, got %v", ev, k, av))
		}
	})
})

// Helper functions

func resourcePtr(value string) *resource.Quantity {
	q := resource.MustParse(value)
	return &q
}

func newMachineTemplate(resources *kubevirtv1.ResourceRequirements, cpu *kubevirtv1.CPU, memory *kubevirtv1.Memory) infrav1.KubevirtMachineTemplate {
	mt := &infrav1.KubevirtMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-template",
			Namespace: "default",
		},
		Spec: infrav1.KubevirtMachineTemplateSpec{
			Template: infrav1.KubevirtMachineTemplateResource{
				Spec: infrav1.KubevirtMachineSpec{
					VirtualMachineTemplate: infrav1.VirtualMachineTemplateSpec{
						Spec: kubevirtv1.VirtualMachineSpec{
							Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
								Spec: kubevirtv1.VirtualMachineInstanceSpec{
									Domain: kubevirtv1.DomainSpec{},
								},
							},
						},
					},
				},
			},
		},
	}

	if resources != nil {
		mt.Spec.Template.Spec.VirtualMachineTemplate.Spec.Template.Spec.Domain.Resources = *resources
	}
	if cpu != nil {
		mt.Spec.Template.Spec.VirtualMachineTemplate.Spec.Template.Spec.Domain.CPU = cpu
	}
	if memory != nil {
		mt.Spec.Template.Spec.VirtualMachineTemplate.Spec.Template.Spec.Domain.Memory = memory
	}

	return *mt
}

