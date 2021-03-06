package servers

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/notifications/application"
	"github.com/onsi/ginkgo"
)

type Notifications struct {
	cmd  *exec.Cmd
	env  application.Environment
	port string
}

func NewNotifications() Notifications {
	return Notifications{
		env: application.NewEnvironment(),
	}
}

func (s Notifications) Compile() {
	path, err := exec.LookPath("go")
	if err != nil {
		panic(err)
	}

	cmd := exec.Cmd{
		Path:   path,
		Args:   []string{"go", "build", "-o", "bin/notifications", "main.go"},
		Dir:    s.env.RootPath,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func (s Notifications) Destroy() {
	err := os.Remove(path.Join(s.env.RootPath, "bin", "notifications"))
	if err != nil {
		panic(err)
	}
}

func (s *Notifications) Boot() {
	s.port = freePort()
	environment := append(os.Environ(), fmt.Sprintf("PORT=%s", s.port))

	cmd := exec.Cmd{
		Path: path.Join(s.env.RootPath, "bin", "notifications"),
		Dir:  s.env.RootPath,
		Env:  environment,
	}
	if os.Getenv("TRACE") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	s.cmd = &cmd

	err := s.cmd.Start()
	if err != nil {
		panic(err)
	}
	s.Ping()
}

func (s Notifications) Ping() {
	timer := time.After(0 * time.Second)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			panic("Failed to boot!")
		case <-timer:
			_, err := http.Get("http://localhost:" + s.port + "/info")
			if err == nil {
				return
			}

			timer = time.After(1 * time.Second)
		}
	}
}

func (s Notifications) Close() {
	err := s.cmd.Process.Kill()
	if err != nil {
		panic(err)
	}
}

func (s *Notifications) Restart() {
	s.Close()
	s.Boot()
}

func (s Notifications) URL() string {
	return "http://localhost:" + s.port
}

func freePort() string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		ginkgo.Fail(err.Error(), 1)
	}
	defer listener.Close()

	address := listener.Addr().String()
	addressParts := strings.SplitN(address, ":", 2)
	return addressParts[1]
}
