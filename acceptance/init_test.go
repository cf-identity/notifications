package acceptance

import (
    "os"
    "os/exec"
    "regexp"
    "testing"

    "github.com/cloudfoundry-incubator/notifications/config"
    "github.com/cloudfoundry-incubator/notifications/gobble"
    "github.com/cloudfoundry-incubator/notifications/models"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var (
    TRUE  = true
    FALSE = false
)

var GUIDRegex = regexp.MustCompile(`[0-9a-f]{8}\-[0-9a-f]{4}\-[0-9a-f]{4}\-[0-9a-f]{4}\-[0-9a-f]{12}`)

func TestAcceptanceSuite(t *testing.T) {
    env := config.NewEnvironment()
    Setup(env)

    RegisterFailHandler(Fail)
    RunSpecs(t, "Acceptance Suite")

    Teardown(env)
}

func Setup(env config.Environment) {
    path, err := exec.LookPath("go")
    if err != nil {
        panic(err)
    }

    cmd := exec.Cmd{
        Path:   path,
        Args:   []string{"go", "build", "-o", "bin/notifications", "main.go"},
        Dir:    env.RootPath,
        Stdout: os.Stdout,
        Stderr: os.Stderr,
    }
    err = cmd.Run()
    if err != nil {
        panic(err)
    }
}

func Teardown(env config.Environment) {
    err := os.Remove(env.RootPath + "/bin/notifications")
    if err != nil {
        panic(err)
    }
}

func TruncateTables() {
    env := config.NewEnvironment()
    models.NewDatabase(env.DatabaseURL).Connection().(*models.Connection).TruncateTables()
    gobble.Database().Connection.TruncateTables()
}
