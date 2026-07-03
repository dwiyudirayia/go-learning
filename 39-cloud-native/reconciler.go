// Modul 39 — Cloud-Native: pola controller/reconcile (Kubernetes), Helm, serverless.
package main

import "fmt"

// POLA CONTROLLER (jantung Kubernetes): loop yang terus-menerus mengamati
// "state aktual" lalu MENYELARASKANNYA dengan "state yang diinginkan" (desired).
// Ini disebut reconcile loop. Operator/controller K8s bekerja persis begini.
//
//	observe (actual) -> diff (desired vs actual) -> act (converge) -> ulangi

// PodSet mensimulasikan resource yang dikelola (mis. Deployment dengan N replika).
type PodSet struct {
	desired int
	actual  int
	events  []string // jejak aksi (untuk demo/test)
}

func NewPodSet(desired int) *PodSet {
	return &PodSet{desired: desired}
}

func (p *PodSet) SetDesired(n int) { p.desired = n }
func (p *PodSet) Actual() int      { return p.actual }
func (p *PodSet) Events() []string { return p.events }

// Reconcile menyelaraskan actual -> desired (scale up/down secukupnya).
// Idempotent: memanggilnya berulang saat sudah selaras tak melakukan apa-apa.
func (p *PodSet) Reconcile() {
	for p.actual < p.desired {
		p.actual++
		p.events = append(p.events, fmt.Sprintf("scale-up -> %d", p.actual))
	}
	for p.actual > p.desired {
		p.actual--
		p.events = append(p.events, fmt.Sprintf("scale-down -> %d", p.actual))
	}
	// Bila actual == desired: tidak ada aksi (steady state).
}
