package main

import "testing"

func TestGLUsable(t *testing.T) {
	cases := []struct {
		version, renderer string
		want              bool
	}{
		{"4.6.0 NVIDIA 551.23", "NVIDIA GeForce RTX 3060", true},
		{"2.1 Mesa 23.0", "SVGA3D", true},
		{"3.3.0 - Build 31.0.101", "Intel(R) UHD Graphics", true},
		{"1.1.0", "GDI Generic", false},
		{"2.0", "Some old card", false},
		{"4.5", "GDI Generic", false},
		{"", "", false},
		{"garbage", "x", false},
	}
	for _, c := range cases {
		if got := glUsable(c.version, c.renderer); got != c.want {
			t.Errorf("glUsable(%q, %q) = %v, want %v", c.version, c.renderer, got, c.want)
		}
	}
}
