package model

import (
	v1 "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/v1/resource"
)

const (
	memoryKey = "memory"
	cpuKey    = "cpu"
)

type Node struct {
	Name string
	Pods []*Pod
}

func (n *Node) FromK8s(node v1.Node) {
	n.Name = node.Name
}

func (n *Node) AddPods(pods []v1.Pod) {
	for _, pod := range pods {
		p := &Pod{}
		p.FromK8s(pod)
		n.Pods = append(n.Pods, p)
	}
}

type Pod struct {
	Name   string
	Memory res.Quantity
	CPU    res.Quantity
}

func (p *Pod) FromK8s(pod v1.Pod) {
	p.Name = pod.Name
	req, _ := resource.PodRequestsAndLimits(&pod)
	p.Memory = req[memoryKey]
	p.CPU = req[cpuKey]
}
