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
	Name string `json:"name"`
	Pods []*Pod `json:"pods"`
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
	Name      string       `json:"name"`
	Memory    res.Quantity `json:"memory"`
	CPU       res.Quantity `json:"cpu"`
	MemoryDec int64        `json:"memory_dec"`
	CPUDec    int64        `json:"cpu_dec"`
}

func (p *Pod) FromK8s(pod v1.Pod) {
	p.Name = pod.Name
	req, _ := resource.PodRequestsAndLimits(&pod)
	p.Memory = req[memoryKey]
	p.CPU = req[cpuKey]
	p.MemoryDec, _ = p.Memory.AsInt64()
	p.CPUDec, _ = p.CPU.AsInt64()
}
