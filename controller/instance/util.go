package instance

import (
	corev1 "k8s.io/api/core/v1"
)

func isNodeReady(node *corev1.Node) bool {
	if node == nil {
		return false
	}
	taints := node.Spec.Taints
	for _, t := range taints {
		if t.Effect == "NoExecute" {
			return false
		}
	}
	return true
}
