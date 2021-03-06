package container

import (
	"reflect"
	"testing"
	"time"

	enginecontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/protobuf/ptypes"
)

func TestVolumesAndBinds(t *testing.T) {
	c := containerConfig{
		task: &api.Task{
			Spec: api.TaskSpec{Runtime: &api.TaskSpec_Container{
				Container: &api.ContainerSpec{
					Mounts: []api.Mount{
						{Type: api.MountTypeBind, Source: "/banana", Target: "/kerfluffle"},
						{Type: api.MountTypeBind, Source: "/banana", Target: "/kerfluffle", BindOptions: &api.Mount_BindOptions{Propagation: api.MountPropagationRPrivate}},
						{Type: api.MountTypeVolume, Source: "banana", Target: "/kerfluffle"},
						{Type: api.MountTypeVolume, Source: "banana", Target: "/kerfluffle", VolumeOptions: &api.Mount_VolumeOptions{NoCopy: true}},
						{Type: api.MountTypeVolume, Target: "/kerfluffle"},
					},
				},
			}},
		},
	}

	config := c.config()
	if len(config.Volumes) != 1 {
		t.Fatalf("expected only 1 anonymous volume: %v", config.Volumes)
	}
	if _, exists := config.Volumes["/kerfluffle"]; !exists {
		t.Fatal("missing anonymous volume entry for target `/kerfluffle`")
	}

	hostConfig := c.hostConfig()
	if len(hostConfig.Binds) != 4 {
		t.Fatalf("exepcted 4 binds: %v", hostConfig.Binds)
	}

	expected := "/banana:/kerfluffle"
	actual := hostConfig.Binds[0]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	expected = "/banana:/kerfluffle:rprivate"
	actual = hostConfig.Binds[1]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	expected = "banana:/kerfluffle"
	actual = hostConfig.Binds[2]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	expected = "banana:/kerfluffle:nocopy"
	actual = hostConfig.Binds[3]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func TestHealthcheck(t *testing.T) {
	c := containerConfig{
		task: &api.Task{
			Spec: api.TaskSpec{Runtime: &api.TaskSpec_Container{
				Container: &api.ContainerSpec{
					Healthcheck: &api.HealthConfig{
						Test:     []string{"a", "b", "c"},
						Interval: ptypes.DurationProto(time.Second),
						Timeout:  ptypes.DurationProto(time.Minute),
						Retries:  10,
					},
				},
			}},
		},
	}
	config := c.config()
	expected := &enginecontainer.HealthConfig{
		Test:     []string{"a", "b", "c"},
		Interval: time.Second,
		Timeout:  time.Minute,
		Retries:  10,
	}
	if !reflect.DeepEqual(config.Healthcheck, expected) {
		t.Fatalf("expected %#v, got %#v", expected, config.Healthcheck)
	}
}

func TestExtraHosts(t *testing.T) {
	c := containerConfig{
		task: &api.Task{
			Spec: api.TaskSpec{Runtime: &api.TaskSpec_Container{
				Container: &api.ContainerSpec{
					Hosts: []string{
						"example.com:1.2.3.4",
						"example.org=5.6.7.8",
						"mylocal 127.0.0.1",
					},
				},
			}},
		},
	}

	hostConfig := c.hostConfig()
	if len(hostConfig.ExtraHosts) != 3 {
		t.Fatalf("exepcted 3 extra hosts: %v", hostConfig.ExtraHosts)
	}

	expected := "example.com:1.2.3.4"
	actual := hostConfig.ExtraHosts[0]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	expected = "example.org:5.6.7.8"
	actual = hostConfig.ExtraHosts[1]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	expected = "mylocal:127.0.0.1"
	actual = hostConfig.ExtraHosts[2]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}
