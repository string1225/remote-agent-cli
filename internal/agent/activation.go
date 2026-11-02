package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

// Keep the proposed credential before redeeming the code. A lost response or
// local write failure can be retried without replacing an existing identity.
func CheckActivationUser(expected string) error {
	if expected == "" {
		return nil
	}
	actual, err := currentProfileOwner()
	if err != nil {
		return err
	}
	if actual != expected {
		return errors.New("激活进程的 Windows 身份与安装用户不一致，请从目标用户桌面运行安装程序；原配置未修改")
	}
	return nil
}

func EnrollActivation(server, code, path string) (Config, error) {
	if err := ValidateServer(server); err != nil {
		return Config{}, err
	}
	server = strings.TrimSuffix(server, "/")
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return Config{}, errors.New("已有本机绑定，请沿用原配置")
	}
	client := &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	// Old servers ignore the proposed-credential field. Check before consuming
	// any code, including when the website has been rolled back during setup.
	capabilities, err := client.Get(server + "/api/activation-capabilities")
	if err != nil {
		return Config{}, errors.New("无法连接激活服务，请检查网络后重试")
	}
	var support struct {
		ProposedCredential bool `json:"proposedCredential"`
	}
	err = json.NewDecoder(io.LimitReader(capabilities.Body, 4096)).Decode(&support)
	capabilities.Body.Close()
	if capabilities.StatusCode != 200 || err != nil || !support.ProposedCredential {
		return Config{}, errors.New("服务器尚未支持 MSI 激活，请先更新服务器；绑定码尚未使用")
	}
	var pending struct{ Server, Code, Credential string }
	journal := path + ".activation-pending.json"
	resumed := false
	if raw, err := os.ReadFile(journal); err == nil {
		resumed = true
		if json.Unmarshal(raw, &pending) != nil || pending.Server != server || code != "" && pending.Code != code {
			return Config{}, errors.New("存在未完成的激活，请使用原绑定码重试；恢复文件已保留")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	} else {
		if len(code) != 43 {
			return Config{}, errors.New("请填写网页生成的设备绑定码")
		}
		pending.Server, pending.Code, pending.Credential = server, code, store.Token()
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return Config{}, err
		}
		raw, _ := json.Marshal(pending)
		if err := writeRecoveryFile(journal, raw); err != nil {
			return Config{}, err
		}
	}
	body, _ := json.Marshal(map[string]string{"token": pending.Code, "credential": pending.Credential, "os": runtime.GOOS, "arch": runtime.GOARCH})
	response, err := client.Post(server+"/api/enroll", "application/json", bytes.NewReader(body))
	if err != nil {
		return Config{}, errors.New("无法连接绑定服务器，已保留激活恢复文件，请重试")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized && !resumed {
		_ = os.Remove(journal)
		return Config{}, errors.New("绑定码无效、已过期或系统类型不匹配，请重新复制绑定码")
	}
	var result struct{ ID, Name, Credential string }
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 16384)).Decode(&result) != nil || result.ID == "" || result.Credential != pending.Credential {
		return Config{}, errors.New("绑定失败：请检查绑定码是否有效、系统类型是否匹配；未完成激活数据已保留")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	root, err := canonicalDir(home)
	if err != nil {
		return Config{}, err
	}
	c := Config{Server: server, ID: result.ID, Name: result.Name, Credential: pending.Credential, AllowedRoots: []string{root}, AllowWrite: true, PermissionMode: "full", Services: []protocol.Service{}, Path: path}
	data, _ := json.MarshalIndent(c, "", "  ")
	if err := writeRecoveryFile(path, data); err != nil {
		return Config{}, errors.New("绑定已接收，但本地配置保存失败；请保留 activation-pending 文件并重试")
	}
	_ = os.Remove(journal)
	return c, nil
}
