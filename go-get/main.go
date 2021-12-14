package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
)

func main() {
	reader, cancel, err := getReader()
	if err != nil {
		log.Fatalf("Get reader failure, nest error: %v\r\n", err)
	}
	defer cancel()

	result, err := parseRepo(reader)
	if err != nil {
		log.Fatalf("Parse repo failure, nest error: %v\r\n", err)
	}

	downloadRepo(result)
}

func downloadRepo(data []string) {
	for _, d := range data {
		var c = fmt.Sprintf("go get -u %s", d)
		_, _, err := ExecuteCmd(c, 10*time.Minute)
		if err != nil {
			fmt.Printf("[%s] %s, nest error: %v\r\n", color.RedString("Failure"), c, err)
		} else {
			fmt.Printf("[%s] %s(download)\r\n", color.GreenString("Success"), c)
		}
	}
}

func parseRepo(reader io.Reader) ([]string, error) {
	var scanner = bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	var result = make([]string, 0, 64)
	for scanner.Scan() {
		var text = scanner.Text()
		if strings.Contains(text, "cannot find package ") && strings.Contains(text, "in any of:") {
			var (
				begin = strings.Index(text, "cannot find package")
				end   = strings.Index(text, "in any of:")
			)

			dot1, dot2 := strings.Index(text, `"`), strings.LastIndex(text, `"`)
			if dot1 == -1 || dot2 == -2 || dot1+1 > dot2 || begin > dot1 || dot2 > end {
				continue
			}
			result = append(result, text[dot1+1:dot2])
		}
	}
	return result, nil
}

func getReader() (io.Reader, func(), error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, nil, err
	}
	if stat == nil {
		return nil, nil, fmt.Errorf("panic: stdin stat is nil")
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, nil, fmt.Errorf("panic: invalid stdin stat mode[%d]", (stat.Mode() & os.ModeCharDevice))
	}
	return os.Stdin, func() {}, nil
}

// ExecuteCmd 执行 command
func ExecuteCmd(c string, timeout time.Duration) (string, string, error) {
	// 超时控制
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var ch = make(chan error)
	var cmd = exec.Command("/bin/sh", "-c", c)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("start execute cmd failure, nest error: %v", err)
	}

	go func() {
		ch <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Signal(syscall.SIGINT)
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-ch
		close(ch)
		return stdout.String(), stderr.String(), fmt.Errorf("execute cmd timeout")

	case err := <-ch:
		close(ch)
		return stdout.String(), stderr.String(), err
	}
}
