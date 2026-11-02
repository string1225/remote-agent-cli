package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func TestActivationRejectsWrongWindowsIdentityBeforeReadingUserState(t *testing.T) {
	t.Setenv("REMOTE_AGENT_ACTIVATION_OWNER", "not-the-current-user")
	var output bytes.Buffer
	if err := runActivation([]string{"--inspect"}, nil, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "身份与安装用户不一致") || strings.Contains(output.String(), "\"bound\":true") {
		t.Fatal(output.String())
	}
}

func TestActivationIndependentChoicesAndPreservedBinding(t *testing.T) {
	for _, options := range []activationOptions{{AutoStart: true}, {InstallCodex: true}, {ConfigureCodex: true, WebProxy: true}} {
		root := t.TempDir()
		path := filepath.Join(root, "config.json")
		c := agent.Config{ID: "device", Name: "PC", Server: "https://example.test/agents", Credential: "private-test-credential", Path: path}
		if err := c.Save(); err != nil {
			t.Fatal(err)
		}
		before, _ := os.ReadFile(path)
		options.Server = c.Server
		options.BindingCode = "unused-code-must-not-rebind"
		var calls []string
		steps := activationSteps{
			Install:   func(agent.Config) error { calls = append(calls, "install"); return nil },
			Configure: func(agent.Config) error { calls = append(calls, "configure"); return nil },
			Service: func(_ agent.Config, automatic bool) error {
				if automatic != options.AutoStart {
					t.Fatal("startup choice lost")
				}
				calls = append(calls, "service")
				return nil
			},
			Proxy: func(_ agent.Config, on bool) error {
				if on != options.WebProxy {
					t.Fatal("proxy choice lost")
				}
				calls = append(calls, "proxy")
				return nil
			},
			Online: func(agent.Config) error { calls = append(calls, "online"); return nil },
		}
		var result activationEvent
		if err := activateDevice(path, options, steps, func(e activationEvent) { result = e }); err != nil {
			t.Fatal(err)
		}
		expected := []string{}
		if options.InstallCodex {
			expected = append(expected, "install")
		}
		if options.ConfigureCodex {
			expected = append(expected, "configure")
		}
		expected = append(expected, "service", "proxy", "online")
		if !reflect.DeepEqual(calls, expected) || !result.Success || !result.Bound {
			t.Fatal(calls, result)
		}
		after, _ := os.ReadFile(path)
		if string(after) != string(before) {
			t.Fatal("original binding overwritten")
		}
		prefs, _ := os.ReadFile(activationPreferences(path))
		if strings.Contains(string(prefs), options.BindingCode) || strings.Contains(string(prefs), c.Credential) {
			t.Fatal("secret saved with choices")
		}
	}
}

func TestActivationPartialFailureRetainsBindingAndAllowsRetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	c := agent.Config{ID: "device", Name: "PC", Server: "https://example.test", Credential: "private-test-credential", Path: path}
	_ = c.Save()
	var serviceCalls int
	steps := activationSteps{Install: func(agent.Config) error { return errors.New("download failed") }, Service: func(agent.Config, bool) error { serviceCalls++; return nil }, Proxy: func(agent.Config, bool) error { return nil }, Online: func(agent.Config) error { return nil }}
	o := activationOptions{Server: c.Server, InstallCodex: true}
	var result activationEvent
	_ = activateDevice(path, o, steps, func(e activationEvent) { result = e })
	if result.Success || !result.Bound || serviceCalls != 1 || !strings.Contains(result.Message, "download failed") {
		t.Fatal(result, serviceCalls)
	}
	steps.Install = func(agent.Config) error { return nil }
	_ = activateDevice(path, o, steps, func(e activationEvent) { result = e })
	if !result.Success {
		t.Fatal(result)
	}
	defaults := activationDefaults(filepath.Join(t.TempDir(), "new.json"))
	if !defaults.AutoStart || !defaults.InstallCodex || !defaults.ConfigureCodex || defaults.WebProxy {
		t.Fatal(defaults)
	}
}
