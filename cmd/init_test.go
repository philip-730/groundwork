package cmd

import "testing"

var registryNameFromURLCases = []struct {
	url  string
	want string
}{
	{"https://github.com/my-org/groundwork-registry", "groundwork-registry"},
	{"https://github.com/my-org/groundwork-registry.git", "groundwork-registry"},
	{"git@github.com:my-org/groundwork-registry", "groundwork-registry"},
	{"git@github.com:my-org/groundwork-registry.git", "groundwork-registry"},
	{"/local/path/to/registry", "registry"},
	{"https://github.com/my-org/groundwork-registry/", "groundwork-registry"},
}

func TestRegistryNameFromURL(t *testing.T) {
	for _, tc := range registryNameFromURLCases {
		t.Run(tc.url, func(t *testing.T) {
			got := registryNameFromURL(tc.url)
			if got != tc.want {
				t.Errorf("registryNameFromURL(%q) = %q, want %q", tc.url, got, tc.want)
			}
		})
	}
}
