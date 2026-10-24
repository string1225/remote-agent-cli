package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLIUserMessagesFollowSelectedLanguage(t *testing.T) {
	for _, language := range []string{"zh", "en", "de-DE"} {
		for _, args := range [][]string{{"codex", "setup", "--help"}, {"unknown-command"}} {
			cmd := exec.Command(os.Args[0], "-test.run=^TestCLIMessageHelper$", "--")
			cmd.Args = append(cmd.Args, args...)
			cmd.Env = append(os.Environ(), "REMOTE_AGENT_CLI_HELPER=1", "REMOTE_AGENT_LANGUAGE="+language)
			out, err := cmd.CombinedOutput()
			if args[0] == "codex" && err != nil {
				t.Fatalf("help: %v %s", err, out)
			}
			if language == "zh" {
				if !strings.Contains(string(out), "用法") && !strings.Contains(string(out), "未知命令") {
					t.Fatalf("Chinese CLI missing: %s", out)
				}
			} else if strings.Contains(string(out), "用法") || strings.Contains(string(out), "未知命令") {
				t.Fatalf("non-Chinese CLI was Chinese: %s", out)
			}
		}
	}
}

func TestCLIMessageHelper(t *testing.T) {
	if os.Getenv("REMOTE_AGENT_CLI_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"remote-agent"}, os.Args[i+1:]...)
			break
		}
	}
	main()
	os.Exit(0)
}
