package main

import "testing"

func TestStatsIncrPort(t *testing.T) {
	statsIncrPort(3744)
	statsIncrPort(3744)
	statsIncrPort(2335)
	for i := 0; i < 125; i++ {
		statsIncrPort(226)
	}
	statsIncrPort(3744)
	statsIncrPort(47223)
	for i := 0; i < 17; i++ {
		statsIncrPort(207)
	}
}
func TestStatsIncrIP(t *testing.T) {
	statsIncrIP("126.235.23.16")
	statsIncrIP("16.63.23.16")
	statsIncrIP("126.235.333.16")
	for i := 0; i < 17; i++ {
		statsIncrIP("126.235.333.16")
	}
	statsIncrIP("126.235.23.16")
	statsIncrIP("126.235.333.16")
	for i := 0; i < 125; i++ {
		statsIncrIP("123.10.315.90")
	}
	statsIncrIP("126.235.333.16")
}

func TestStatsGetNUniquePorts(t *testing.T) {}
func TestStatsGetNUniqueIPs(t *testing.T)   {}

func TestStatsGetLeadingPorts(t *testing.T) {
	TestStatsIncrPort(t)
	ls := statsGetLeadingPorts(3)
	t.Log(ls)
}
func TestStatsGetLeadingIPs(t *testing.T) {
	TestStatsIncrIP(t)
	ls := statsGetLeadingIPs(5)
	t.Log(ls)
}

func TestStatsClearPorts(t *testing.T) {}
func TestStatsClearIPs(t *testing.T)   {}

func TestLeadersStringers(t *testing.T) {
	TestStatsIncrPort(t)
	lsports := statsGetLeadingPorts(3)

	TestStatsIncrIP(t)
	lsips := statsGetLeadingIPs(5)

	txtLeadersPorts := lsports.String()
	txtLeadersIPs := lsips.String()

	t.Log(txtLeadersPorts, txtLeadersIPs)
}
