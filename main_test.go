package main

import "testing"

func TestGetTags(t *testing.T) {
	tc := []struct {
		port     int
		wanttags string
	}{{
		port:     0,
		wanttags: "",
	}, {
		port:     6379,
		wanttags: "#redis #vital",
	}, {
		port:     80,
		wanttags: "#http",
	}, {
		port:     235,
		wanttags: "",
	}, {
		port:     42023,
		wanttags: "",
	}, {
		port:     8325,
		wanttags: "",
	}, {
		port:     443,
		wanttags: "#https",
	}}

	for i, tcase := range tc {
		if actualTags := getTags(Itoa(tcase.port)); actualTags != tcase.wanttags {
			t.Logf("Got wrong tags in test #%d. Want %s, got %s", i, tcase.wanttags, actualTags)
			t.Fail()
		}
	}
}
