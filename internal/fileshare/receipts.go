package fileshare

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

// Management proofs stay local. They are never included in CLI/MCP output,
// public URLs, platform requests, or file metadata shared with recipients.
type receipt struct {
	File    protocol.SharedFile `json:"file"`
	Visitor string              `json:"visitor"`
	Owner   string              `json:"owner"`
}

var visitorPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// Atomically publish a complete identity with a create-only hard link. Racing
// uploads reuse the winner; a restart never rotates identity to evade quotas.
func visitorIdentity(dir string) (string, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	read := func() (string, error) {
		info, e := root.Lstat(".visitor-token")
		if e != nil {
			return "", e
		}
		if !info.Mode().IsRegular() || info.Size() != 43 {
			return "", errors.New("临时文件身份记录无效，已保留原记录")
		}
		f, e := root.Open(".visitor-token")
		if e != nil {
			return "", e
		}
		defer f.Close()
		raw, e := io.ReadAll(io.LimitReader(f, 44))
		if e != nil || !visitorPattern.Match(raw) {
			return "", errors.New("临时文件身份记录无效，已保留原记录")
		}
		return string(raw), nil
	}
	if token, e := read(); e == nil {
		return token, nil
	} else if !errors.Is(e, os.ErrNotExist) {
		return "", e
	}
	token := randomName()
	temporary := ".identity-" + randomName()
	f, err := root.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer root.Remove(temporary)
	_, err = f.WriteString(token)
	if err == nil {
		err = f.Sync()
	}
	closed := f.Close()
	if err != nil {
		return "", err
	}
	if closed != nil {
		return "", closed
	}
	if err = root.Link(temporary, ".visitor-token"); err != nil && !errors.Is(err, os.ErrExist) {
		return "", errors.New("无法原子保存临时文件身份，尚未上传")
	}
	return read()
}

func prepareReceipts(dir string) error {
	if !filepath.IsAbs(dir) {
		return errors.New("管理回执目录必须是绝对路径")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.New("无法准备临时文件管理回执目录；尚未上传")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return errors.New("无法打开管理回执目录")
	}
	defer root.Close()
	// Check write permission before allocating remote storage. This file contains
	// no secret and is removed immediately.
	name := ".probe-" + randomName()
	f, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("管理回执目录不可写；尚未上传")
	}
	err = f.Close()
	_ = root.Remove(name)
	return err
}

func saveReceipt(dir string, r receipt) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	temporary := ".receipt-" + randomName()
	f, err := root.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer root.Remove(temporary)
	err = json.NewEncoder(f).Encode(r)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if _, err = root.Lstat(r.File.ID + ".json"); !errors.Is(err, os.ErrNotExist) {
		return errors.New("回执已存在或不可检查")
	}
	return root.Link(temporary, r.File.ID+".json")
}

func (c *Client) Delete(ctx context.Context, id, dir string) error {
	if !fileID.MatchString(id) {
		return errors.New("只支持删除本机有回执的storage.to九位文件ID")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return errors.New("本机没有该文件的管理回执，不能删除")
	}
	defer root.Close()
	info, err := root.Lstat(id + ".json")
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("管理回执不存在或不是普通文件")
	}
	f, err := root.Open(id + ".json")
	if err != nil {
		return errors.New("无法读取管理回执")
	}
	var r receipt
	err = json.NewDecoder(io.LimitReader(f, 8192)).Decode(&r)
	f.Close()
	if err != nil || r.File.Provider != "storage.to" || r.File.ID != id || r.File.URL != "https://storage.to/"+id || r.Owner == "" {
		return errors.New("临时文件管理回执无效")
	}
	if err = c.api(ctx, "DELETE", "/file/"+id, nil, r.Visitor, r.Owner, nil); err != nil {
		return err
	}
	return root.Remove(id + ".json")
}
