// Modul 39 — Cloud-Native: pola controller/reconcile (Kubernetes), Helm, serverless.
package main

import "fmt"

// 🔍 Analogi besar reconcile loop: seperti TERMOSTAT AC. Kamu set suhu yang DIINGINKAN (desired,
// mis. 24°C). Termostat terus-menerus membandingkan dengan suhu AKTUAL ruangan, lalu bertindak:
// terlalu panas -> dinginkan; terlalu dingin -> hangatkan; pas -> diam. Kubernetes bekerja persis
// begini: kamu bilang "aku mau 3 replika", lalu controller terus menyelaraskan realita ke keinginan
// itu — ada pod mati? buat baru. Kelebihan? matikan. Kamu deklarasikan TUJUAN, bukan langkah manual.
//
// 🔍 Analogi idempotent (lagi): memanggil Reconcile saat sudah pas = menekan tombol lift di lantai
// yang sedang kamu tempati -> tak terjadi apa-apa. Aman dipanggil berkali-kali. Ini inti model "deklaratif".

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
