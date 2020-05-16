package dbtest

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"
)

// Container tracks information about a docker container started for tests.
type Container struct {
	ID   string
	Host string // IP:Port
}

// Start docker container for test session.
func Start(t *testing.T) *Container {
	t.Helper()

	// > docker run -P -d postgres:11.1-alpine
	//   009362afa6b043819098cd5d12aa4beba12665f21d050ad2b736863a091324d9
	//cmd := exec.Command("docker", "run", "-P", "-d", "postgres:11.1-alpine")
	cmd := exec.Command("docker", "run", "-p", "5555:5432", "-d", "postgres:11.1-alpine")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container %v", err)
	}

	id := out.String()[:12]
	t.Log("DB ContainerID:", id)

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
	}

	var doc []struct {
		NetworkSettings struct {
			Ports struct {
				TCP5432 []struct {
					HostIP   string `json:"HostIp"`
					HostPort string `json:"HostPort"`
				} `json:"5432/tcp"`
			} `json:"Ports"`
		} `json:"NetworkSettings"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}

	network := doc[0].NetworkSettings.Ports.TCP5432[0]

	c := Container{
		ID:   id,
		Host: network.HostIP + ":" + network.HostPort,
	}

	t.Log("DB Host:", c.Host)

	return &c
}

// StopContainer stops and removes the specified container.
func StopContainer(t *testing.T, c *Container) {
	t.Helper()

	if err := exec.Command("docker", "stop", c.ID).Run(); err != nil {
		t.Fatalf("could not stop container: %v", err)
	}
	t.Log("Stopped:", c.ID)

	if err := exec.Command("docker", "rm", c.ID, "-v").Run(); err != nil {
		t.Fatalf("could not remove container: %v", err)
	}
	t.Log("Removed:", c.ID)
}

// DumpContainerLogs runs "docker logs" against the container and send it to t.Log
func DumpContainerLogs(t *testing.T, c *Container) {
	t.Helper()

	out, err := exec.Command("docker", "logs", c.ID).CombinedOutput()
	if err != nil {
		t.Fatalf("could not log container: %v", err)
	}
	t.Logf("Logs for %s\n%s:", c.ID, out)
}
