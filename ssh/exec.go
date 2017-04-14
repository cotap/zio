package ssh

import (
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/goware/prefixer"
)

var colors []color.Attribute

func init() {
	colors = []color.Attribute{
		color.FgRed,
		color.FgGreen,
		color.FgYellow,
		color.FgBlue,
		color.FgMagenta,
		color.FgCyan,
	}
}

func SSH(ipAddress string, sshCmd string) error {
	if ipAddress == "" {
		return errors.New("No IP address")
	}

	var cmd *exec.Cmd
	if sshCmd != "" {
		sshCmd = strings.Replace(sshCmd, "'", "\\'", -1)
		cmd = exec.Command("/usr/bin/env", "ssh", ipAddress, "-t", "exec $SHELL -l -c '"+sshCmd+"'")
	} else {
		cmd = exec.Command("/usr/bin/env", "ssh", ipAddress, "-t")
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func Exec(ipAddress, command string) error {
	if ipAddress == "" {
		return errors.New("No IP address")
	}

	cmd := exec.Command("/usr/bin/env", "ssh", ipAddress, "-tt", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	cAttr := colors[hash(ipAddress)%len(colors)]
	prefix := color.New(cAttr).Sprint(fmt.Sprintf("%-15s| ", ipAddress))
	prefixedStdout := prefixer.New(stdout, prefix)
	prefixedStdErr := prefixer.New(stderr, prefix)

	go prefixedStdout.WriteTo(os.Stdout)
	go prefixedStdErr.WriteTo(os.Stderr)

	return cmd.Run()
}

func ExecAll(ipAddresses []string, command string, concurrency int) {
	pool := make(chan bool, concurrency)

	// fill pool and wait for draining before continuing
	for _, ipAddress := range ipAddresses {
		pool <- true
		go func(ipAddress string) {
			defer func() { <-pool }()
			if err := Exec(ipAddress, command); err != nil {
				log.Println(err)
			}
		}(ipAddress)
	}

	// fill pool to make sure all have completed
	for i := 0; i < cap(pool); i++ {
		pool <- true
	}
}

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
