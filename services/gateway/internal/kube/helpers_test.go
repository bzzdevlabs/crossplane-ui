package kube_test

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// metav1ListOptions is a typed helper so test files can stay free of the
// verbose k8s import every time they call List.
func metav1ListOptions() metav1.ListOptions { return metav1.ListOptions{} }
