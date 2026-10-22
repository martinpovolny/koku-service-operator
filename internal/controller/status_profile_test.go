package controller

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	costv1alpha1 "github.com/project-koku/koku-service-operator/api/v1alpha1"
)

func TestReconcileSetsStatusProfile(t *testing.T) {
	cases := []struct {
		name    string
		profile costv1alpha1.Profile
		paused  bool
		want    costv1alpha1.Profile
	}{
		{"unset defaults to small", "", false, costv1alpha1.ProfileSmall},
		{"large", costv1alpha1.ProfileLarge, false, costv1alpha1.ProfileLarge},
		{"unset paused", "", true, costv1alpha1.ProfileSmall},
		{"xlarge paused", costv1alpha1.ProfileXLarge, true, costv1alpha1.ProfileXLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cr := minimalCR(testCRName, testNamespace)
			controllerutil.AddFinalizer(cr, finalizerName)
			cr.Spec.Profile = tc.profile
			if tc.paused {
				cr.Annotations = map[string]string{pauseAnnotation: annotationTrue}
			}
			c := fake.NewClientBuilder().
				WithScheme(ownershipScheme(t)).
				WithObjects(cr).
				WithStatusSubresource(cr).
				Build()
			r := &CostManagementServiceConfigReconciler{
				Client:   c,
				Scheme:   ownershipScheme(t),
				Recorder: record.NewFakeRecorder(100),
			}
			key := types.NamespacedName{Name: testCRName, Namespace: testNamespace}
			// Phases may fail against the fake client; status is patched regardless.
			_, _ = r.Reconcile(context.Background(), reconcile.Request{NamespacedName: key})

			got := &costv1alpha1.CostManagementServiceConfig{}
			if err := c.Get(context.Background(), key, got); err != nil {
				t.Fatalf("get CR: %v", err)
			}
			if got.Status.Profile != tc.want {
				t.Errorf("status.profile = %q, want %q", got.Status.Profile, tc.want)
			}
		})
	}
}
