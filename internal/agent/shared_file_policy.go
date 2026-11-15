package agent

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/string1225/remote-agent-cli/internal/fileshare"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

const FileSharingInstructions = "文件分享分为temporary和sensitive：普通代码包、生成产物等默认temporary，使用storage.to客户端直传，返回下载页面和实际expiresAt，不是直链。特别敏感文件须先取得用户针对该文件和分享范围的明确确认，再用class=sensitive、userConfirmed=true走原139通道；该确认参数不能替代询问，不因失败擅自填写。139仍是持链接可下载、下次北京时间凌晨4点失效，不是私密保险箱。不要将凭据、密钥、真实环境文件混入普通包；上传前核对范围。storage.to不可用、拒收类型或额度用尽时报告具体错误，不自动改走139或改名规避拒收。不迁移旧文件。upload_file不发送消息；delete_uploaded_file仅删除本机留有管理回执的storage.to副本，保留原文件。"

type FileUploadOptions struct {
	Class         string
	UserConfirmed bool
}

func (o FileUploadOptions) Validate() error {
	if o.Class != "" && o.Class != "temporary" && o.Class != "sensitive" {
		return errors.New("文件分类只能是 temporary 或 sensitive")
	}
	if o.Class == "sensitive" && !o.UserConfirmed {
		return errors.New("敏感文件尚未获得用户确认；请先确认该文件、分享范围及139持链接可下载和凌晨4点失效，再设置确认参数")
	}
	return nil
}

// UploadFile defaults to the public temporary provider. No device credential is
// sent there; the enrolled site is used only through an explicitly confirmed
// sensitive request. Provider failures never change the selected route.
func UploadFile(ctx context.Context, c Config, path string) (protocol.SharedFile, error) {
	return UploadFileWithOptions(ctx, c, path, FileUploadOptions{})
}

func UploadFileWithOptions(ctx context.Context, c Config, path string, o FileUploadOptions) (protocol.SharedFile, error) {
	if err := o.Validate(); err != nil {
		return protocol.SharedFile{}, err
	}
	if o.Class == "sensitive" {
		result, err := uploadFile139(ctx, c, path)
		if err == nil {
			result.Provider = "139"
			result.Class = "sensitive"
			result.LinkType = "direct"
		}
		return result, err
	}
	return fileshare.New(nil).Upload(ctx, path, fileReceiptDir(c))
}

func fileReceiptDir(c Config) string {
	path := c.Path
	if path == "" {
		path = DefaultPath()
	}
	if absolute, err := filepath.Abs(path); err == nil {
		path = absolute
	}
	return filepath.Join(filepath.Dir(path), "shared-file-receipts")
}

func DeleteUploadedFile(ctx context.Context, c Config, id string) error {
	return fileshare.New(nil).Delete(ctx, id, fileReceiptDir(c))
}
