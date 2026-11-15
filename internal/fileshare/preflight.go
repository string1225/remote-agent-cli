package fileshare

import (
	"archive/zip"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/store"
)

func randomName() string { return store.Token() }

func credentialName(name string) bool {
	name = strings.ToLower(path.Base(strings.ReplaceAll(name, `\`, "/")))
	if name == ".env" || name == "id_rsa" || name == "id_ed25519" || name == "auth.json" || name == "credentials.json" {
		return true
	}
	if strings.HasPrefix(name, ".env.") && !strings.HasSuffix(name, ".example") && !strings.HasSuffix(name, ".sample") && !strings.HasSuffix(name, ".template") {
		return true
	}
	return strings.HasSuffix(name, ".key") || strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".p12") || strings.HasSuffix(name, ".pfx")
}

func checkTemporaryFile(f *os.File, name string, size int64) error {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".exe", ".scr", ".com", ".pif", ".apk", ".ipa":
		return errors.New("storage.to不接受该文件类型；不会改名规避或自动转139，请用户选择分享方式")
	}
	if credentialName(name) {
		return errors.New("文件名疑似凭据或密钥，不能作为普通临时文件上传；请先按敏感文件确认分享范围")
	}
	if strings.EqualFold(filepath.Ext(name), ".zip") {
		archive, err := zip.NewReader(f, size)
		if err != nil {
			return errors.New("ZIP目录无法检查，请核对文件后重试")
		}
		if len(archive.File) > 100000 {
			return errors.New("ZIP目录超过检查上限，请缩小分享范围")
		}
		for _, entry := range archive.File {
			if credentialName(entry.Name) {
				return errors.New("ZIP包含疑似凭据或密钥文件名，不能作为普通临时包上传；请排除敏感文件或先取得用户确认")
			}
		}
	}
	return nil
}
