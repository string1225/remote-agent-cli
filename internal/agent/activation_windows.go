package agent

import (
	"errors"
	"os"
)

func ActivationConfigPath() (string, error) {
	profile, err := ReadServiceProfile(serviceProfilePath())
	if errors.Is(err, os.ErrNotExist) {
		return DefaultPath(), nil
	}
	if err != nil {
		return "", errors.New("原 Windows Service 配置无法读取，请先恢复原 service.json；未创建新绑定")
	}
	owner, err := currentProfileOwner()
	if err != nil {
		return "", err
	}
	if profile.OwnerSID != "" && profile.OwnerSID != owner {
		return "", errors.New("本机服务属于其他 Windows 用户，请使用原设备用户运行安装程序")
	}
	return profile.Config, nil
}
