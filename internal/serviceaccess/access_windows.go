package serviceaccess

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const maxObjects = 20000

type object struct {
	Path      string `json:"path"`
	Identity  string `json:"identity"`
	Directory bool   `json:"directory"`
	Before    string `json:"before"`
	After     string `json:"after,omitempty"`
	handle    windows.Handle
}

type backup struct {
	Version    int      `json:"version"`
	ID         string   `json:"id"`
	Created    string   `json:"created"`
	Status     string   `json:"status"` // prepared, active, restored; prepared also covers interrupted transactions.
	Path       string   `json:"path"`
	Permission string   `json:"permission"`
	Recursive  bool     `json:"recursive"`
	GitTrust   bool     `json:"gitTrust"`
	SID        string   `json:"sid"`
	Objects    []object `json:"objects"`
}

type manager struct {
	base, sid string
	// Tests inject faults and a private base; they never access production state,
	// SCM, real UAC, or credentials.
	fault func(string) error
	pins  []windows.Handle
}

func trustPath() string {
	root, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0)
	if err != nil {
		return ""
	}
	return filepath.Join(root, "RemoteAgent", "access", "trust.gitconfig")
}

func Execute(o Options, w io.Writer) error {
	if err := o.Validate(); err != nil {
		return err
	}
	// Read-only SCM inspection. Never register, restart, or change a service.
	scmHandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return err
	}
	scm := &mgr.Mgr{Handle: scmHandle}
	defer scm.Disconnect()
	name, _ := windows.UTF16PtrFromString("RemoteAgent")
	serviceHandle, err := windows.OpenService(scm.Handle, name, windows.SERVICE_QUERY_CONFIG)
	if err != nil {
		return fmt.Errorf("请先安装 RemoteAgent Windows Service：%w", err)
	}
	s := &mgr.Service{Name: "RemoteAgent", Handle: serviceHandle}
	defer s.Close()
	c, err := s.Config()
	if err != nil {
		return err
	}
	if c.SidType != windows.SERVICE_SID_TYPE_UNRESTRICTED || !strings.EqualFold(c.ServiceStartName, `NT AUTHORITY\LocalService`) {
		return errors.New("服务必须使用原 LocalService 身份和已启用的 RemoteAgent 服务 SID；本命令不会更改服务身份或重启服务")
	}
	sid, _, _, err := windows.LookupSID("", `NT SERVICE\RemoteAgent`)
	if err != nil {
		return err
	}
	base := filepath.Dir(trustPath())
	if base == "." {
		return errors.New("无法确定 ProgramData 目录")
	}
	m := &manager{base: base, sid: sid.String()}
	defer m.close()
	if o.Action == "list" {
		if _, err = os.Stat(base); errors.Is(err, fs.ErrNotExist) {
			_, err = fmt.Fprintln(w, "尚无 access 授权备份。")
			return err
		}
		return m.list(w)
	}
	if !windows.GetCurrentProcessToken().IsElevated() {
		return errors.New("修改后台访问权限需要管理员授权，请从 remote-agent access 命令触发 UAC")
	}
	if err = m.prepare(); err != nil {
		return err
	}
	// Serialize all grants/restores, including different CLI processes.
	lock, err := windows.UTF16PtrFromString(filepath.Join(base, ".lock"))
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(lock, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_ALWAYS, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return fmt.Errorf("另一个 access 操作正在进行或无法锁定备份目录：%w", err)
	}
	defer windows.CloseHandle(h)
	return m.run(o, w)
}

func (m *manager) close() {
	for i := len(m.pins) - 1; i >= 0; i-- {
		_ = windows.CloseHandle(m.pins[i])
	}
	m.pins = nil
}

func (m *manager) prepare() error {
	// Pin ancestors against rename and reject links before creating our private
	// directory. Only this new directory gets a replacement DACL.
	if _, err := m.open(filepath.Dir(m.base), false); err != nil {
		return err
	}
	sd, err := windows.SecurityDescriptorFromString("O:BAG:BAD:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;FRFX;;;" + m.sid + ")")
	if err != nil {
		return err
	}
	p, _ := windows.UTF16PtrFromString(m.base)
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	err = windows.CreateDirectory(p, &sa)
	if err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return err
	}
	if _, err = m.open(m.base, false); err != nil {
		return err
	}
	// An existing store must already have the intended protected ACL/owner. Do
	// not silently take over a user-created or redirected state directory.
	actual, err := windows.GetNamedSecurityInfo(m.base, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := actual.Owner()
	if err != nil || owner.String() != "S-1-5-32-544" {
		return errors.New("授权备份目录属主异常；请由管理员检查 ProgramData/RemoteAgent/access，未修改目标权限")
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	// Normalize only this dedicated store. Never modify RemoteAgent's parent.
	return windows.SetNamedSecurityInfo(m.base, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
}

func (m *manager) run(o Options, w io.Writer) error {
	if o.Action == "restore" {
		return m.restore(o, w)
	}
	return m.grant(o, w)
}

func (m *manager) checkpoint(stage string) error {
	if m.fault != nil {
		return m.fault(stage)
	}
	return nil
}

// open pins every ancestor and the final object without FILE_SHARE_DELETE.
// A checked path therefore cannot be renamed into a junction during mutation.
func (m *manager) open(path string, write bool) (object, error) {
	path = filepath.Clean(path)
	root := filepath.VolumeName(path) + `\`
	parts := strings.Split(strings.TrimPrefix(path, root), `\`)
	cursor := strings.TrimRight(root, `\`)
	sequence := []string{root}
	for _, part := range parts {
		if part != "" {
			cursor += `\` + part
			sequence = append(sequence, cursor)
		}
	}
	var result object
	for i, name := range sequence {
		access := uint32(windows.FILE_READ_ATTRIBUTES)
		if i == len(sequence)-1 {
			access |= windows.READ_CONTROL
			if write {
				access |= windows.WRITE_DAC
			}
		}
		p, err := windows.UTF16PtrFromString(name)
		if err != nil {
			return result, err
		}
		h, err := windows.CreateFile(p, access, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
		if err != nil {
			return result, fmt.Errorf("无法打开 %s（未接管属主或绕过拒绝权限）：%w", name, err)
		}
		m.pins = append(m.pins, h)
		var info windows.ByHandleFileInformation
		if err = windows.GetFileInformationByHandle(h, &info); err != nil {
			return result, err
		}
		if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			return result, fmt.Errorf("拒绝重解析点（符号链接/junction）：%s", name)
		}
		if i == len(sequence)-1 {
			result = object{Path: path, Directory: info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0, Identity: fmt.Sprintf("%x:%x:%x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow), handle: h}
			if !result.Directory && info.NumberOfLinks > 1 {
				return result, fmt.Errorf("拒绝多硬链接文件，避免授权所选路径之外的对象：%s", name)
			}
			// Normalize 8.3 aliases/casing through the pinned handle, so aliases
			// cannot bypass overlap checks or create a second Git trust spelling.
			buffer := make([]uint16, 32768)
			n, err := windows.GetFinalPathNameByHandle(h, &buffer[0], uint32(len(buffer)), 0)
			if err != nil {
				return result, fmt.Errorf("无法规范化路径 %s：%w", name, err)
			}
			if n >= uint32(len(buffer)) {
				return result, fmt.Errorf("规范化路径过长：%s", name)
			}
			result.Path = strings.TrimPrefix(windows.UTF16ToString(buffer[:n]), `\\?\`)
			if err = validatePath(result.Path); err != nil {
				return result, err
			}
		}
	}
	return result, nil
}

func descriptor(o object) (string, error) {
	sd, err := windows.GetSecurityInfo(o.handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION)
	if err != nil {
		return "", err
	}
	return sd.String(), nil
}

func setDACL(o object, sddl string) error {
	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	control, _, err := sd.Control()
	if err != nil {
		return err
	}
	flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION)
	if control&windows.SE_DACL_PROTECTED != 0 {
		flags = windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION
	}
	err = windows.SetSecurityInfo(o.handle, windows.SE_FILE_OBJECT, flags, nil, nil, dacl, nil)
	runtime.KeepAlive(sd)
	return err
}

func comparableSDDL(sddl string) string {
	// Windows normalizes SE_DACL_AUTO_INHERITED when propagating inheritance.
	// This bookkeeping bit is not a permission/protection change. Keep every
	// ACE, owner/group and SE_DACL_PROTECTED bit in comparisons.
	start := strings.Index(sddl, "D:")
	if start < 0 {
		return sddl
	}
	end := strings.Index(sddl[start:], "(")
	if end < 0 {
		end = len(sddl) - start
	}
	end += start
	return sddl[:start] + strings.ReplaceAll(sddl[start:end], "AI", "") + sddl[end:]
}

func grantDACL(o object, sid, permission string, recursive bool) error {
	sd, err := windows.GetSecurityInfo(o.handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	previous, _, err := sd.DACL()
	if err != nil || previous == nil {
		return errors.New("拒绝缺失/NULL DACL；不会把不受限 ACL 改成受限 ACL")
	}
	identity, err := windows.StringToSid(sid)
	if err != nil {
		return err
	}
	var pin runtime.Pinner
	pin.Pin(identity)
	defer pin.Unpin()
	rights := windows.ACCESS_MASK(windows.FILE_GENERIC_READ | windows.FILE_GENERIC_EXECUTE)
	if permission == "modify" {
		rights |= windows.FILE_GENERIC_WRITE | windows.DELETE
	}
	var inheritance uint32
	if recursive && o.Directory {
		inheritance = windows.CONTAINER_INHERIT_ACE | windows.OBJECT_INHERIT_ACE
	}
	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{AccessPermissions: rights, AccessMode: windows.GRANT_ACCESS, Inheritance: inheritance, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeType: windows.TRUSTEE_IS_WELL_KNOWN_GROUP, TrusteeValue: windows.TrusteeValueFromSID(identity)}}}, previous)
	if err != nil {
		return err
	}
	err = windows.SetSecurityInfo(o.handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
	runtime.KeepAlive(sd)
	runtime.KeepAlive(identity)
	runtime.KeepAlive(acl)
	return err
}

func overlaps(a, b string) bool {
	a, b = strings.ToLower(filepath.Clean(a)), strings.ToLower(filepath.Clean(b))
	return a == b || strings.HasPrefix(a, b+`\`) || strings.HasPrefix(b, a+`\`)
}

func (m *manager) records() ([]backup, error) {
	entries, err := os.ReadDir(m.base)
	if err != nil {
		return nil, err
	}
	var result []backup
	for _, entry := range entries {
		id := strings.TrimSuffix(entry.Name(), ".json")
		if entry.Name() != id+".json" || !backupID.MatchString(id) {
			continue
		}
		p := filepath.Join(m.base, entry.Name())
		if info, err := os.Lstat(p); err != nil || !info.Mode().IsRegular() || info.Size() > 128<<20 {
			return nil, fmt.Errorf("备份文件异常：%s", p)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var b backup
		if err = json.Unmarshal(raw, &b); err != nil {
			return nil, fmt.Errorf("备份损坏，未修改目标：%s：%w", p, err)
		}
		if b.Version != 1 || b.ID != id || b.SID != m.sid || (b.Status != "prepared" && b.Status != "active" && b.Status != "restored") || len(b.Objects) == 0 || len(b.Objects) > maxObjects {
			return nil, fmt.Errorf("备份格式异常：%s", p)
		}
		result = append(result, b)
	}
	return result, nil
}

func (m *manager) put(name string, raw []byte) error {
	if err := m.checkpoint("write:" + name); err != nil {
		return err
	}
	f, err := os.CreateTemp(m.base, ".access-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	_, err = f.Write(raw)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	from, _ := windows.UTF16PtrFromString(tmp)
	to, _ := windows.UTF16PtrFromString(filepath.Join(m.base, name))
	return windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}

func (m *manager) save(b backup) error {
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return m.put(b.ID+".json", raw)
}

func (m *manager) publish(records []backup) error {
	var paths []string
	for _, b := range records {
		if b.Status == "active" && b.GitTrust {
			paths = append(paths, b.Path)
		}
	}
	sort.Strings(paths)
	config, err := gitConfig(paths)
	if err != nil {
		return err
	}
	return m.put("trust.gitconfig", []byte(config))
}

func (m *manager) list(w io.Writer) error {
	if _, err := m.open(m.base, false); err != nil {
		return err
	}
	records, err := m.records()
	if err != nil {
		return err
	}
	for _, b := range records {
		if _, err = fmt.Fprintf(w, "%s  %s  %s  %s  recursive=%t git-trust=%t\n", b.ID, b.Status, b.Permission, b.Path, b.Recursive, b.GitTrust); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(w, "备份目录：%s（不含凭据；备份保留供恢复）\n", m.base)
	return err
}

func (m *manager) grant(o Options, w io.Writer) error {
	root, err := m.open(o.Path, true)
	if err != nil {
		return err
	}
	o.Path = root.Path
	store, err := m.open(m.base, false)
	if err != nil {
		return err
	}
	// A grant may never make its own protected state writable by the service.
	if overlaps(o.Path, filepath.Dir(store.Path)) {
		return errors.New("不能授权 RemoteAgent 系统状态目录或其祖先；请选择业务目录")
	}
	records, err := m.records()
	if err != nil {
		return err
	}
	for _, b := range records {
		if b.Status != "restored" && overlaps(b.Path, o.Path) {
			return fmt.Errorf("与备份 %s 的授权重叠；请先 restore，避免互相覆盖", b.ID)
		}
	}
	var token [16]byte
	if _, err = rand.Read(token[:]); err != nil {
		return err
	}
	b := backup{Version: 1, ID: hex.EncodeToString(token[:]), Created: time.Now().UTC().Format(time.RFC3339), Status: "prepared", Path: filepath.Clean(o.Path), Permission: o.Permission, Recursive: o.Recursive, GitTrust: o.GitTrust, SID: m.sid}
	err = filepath.WalkDir(b.Path, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if len(b.Objects) >= maxObjects {
			return fmt.Errorf("一次授权最多 %d 个对象；请缩小路径范围", maxObjects)
		}
		obj, err := m.open(path, true)
		if err != nil {
			return err
		}
		obj.Before, err = descriptor(obj)
		if err != nil {
			return err
		}
		b.Objects = append(b.Objects, obj)
		if !o.Recursive && d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return err // Preflight is complete before any target DACL changes.
	}
	if o.Recursive && !b.Objects[0].Directory {
		return errors.New("--recursive 只适用于目录")
	}
	if o.GitTrust {
		if err = m.checkRepository(b.Objects[0]); err != nil {
			return err
		}
	}
	if err = m.save(b); err != nil {
		return err
	}
	fmt.Fprintf(w, "已保存权限备份：%s\n恢复命令：remote-agent access restore --backup %s\n", b.ID, b.ID)
	rollback := func(cause error) error {
		// Restore parent first so Windows removes the introduced inheritance;
		// restore children afterwards to preserve originally protected DACLs.
		var recovery error
		for _, obj := range b.Objects {
			recovery = errors.Join(recovery, setDACL(obj, obj.Before))
		}
		recovery = errors.Join(recovery, m.publish(records))
		if recovery == nil {
			b.Status = "restored"
			recovery = m.save(b)
		}
		if recovery != nil {
			return errors.Join(cause, fmt.Errorf("自动恢复未完成：%w；备份 %s 已保留，请用 access restore --backup %s --force 恢复", recovery, b.ID, b.ID))
		}
		return fmt.Errorf("授权失败，已恢复原 DACL 和 Git 信任；备份 %s 保留：%w", b.ID, cause)
	}
	for _, obj := range b.Objects {
		if err = m.checkpoint("grant:" + obj.Path); err == nil {
			err = grantDACL(obj, m.sid, o.Permission, o.Recursive)
		}
		if err != nil {
			return rollback(err)
		}
	}
	for i := range b.Objects {
		b.Objects[i].After, err = descriptor(b.Objects[i])
		if err != nil {
			return rollback(err)
		}
	}
	b.Status = "active"
	if err = m.save(b); err != nil {
		return rollback(err)
	}
	if err = m.publish(append(records, b)); err != nil {
		return rollback(err)
	}
	_, err = fmt.Fprintf(w, "已仅向 RemoteAgent 服务 SID %s 授予 %s：%s\nGit 仓库信任=%t；GitHub 登录继续使用 git login / github login。\n", m.sid, o.Permission, b.Path, o.GitTrust)
	return err
}

func (m *manager) checkRepository(root object) error {
	if !root.Directory {
		return errors.New("--git-trust 必须选择仓库根目录")
	}
	// Do not execute Git or read repository configuration before trust exists.
	// Reject linked worktrees here: their separate gitdir needs its own explicit
	// ACL grant and cannot be silently pulled into this authorization.
	marker, err := m.open(filepath.Join(root.Path, ".git"), false)
	if err == nil && marker.Directory {
		return nil
	}
	return errors.New("--git-trust 路径必须包含真实 .git 目录；linked worktree/裸仓库请分别管理 Git 目录和信任")
}

func (m *manager) restore(o Options, w io.Writer) error {
	records, err := m.records()
	if err != nil {
		return err
	}
	index := -1
	for i := range records {
		if records[i].ID == o.Backup {
			index = i
		}
	}
	if index < 0 {
		return errors.New("未找到授权备份；请运行 access list")
	}
	b := &records[index]
	if b.Status == "restored" {
		// Reconcile an interrupted final trust publication; never reapply old ACLs.
		if err = m.publish(records); err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, "该备份已恢复，原备份仍保留。")
		return err
	}
	if b.Status == "prepared" && !o.Force {
		return errors.New("这是中断/未完成的授权；请核对备份后使用 restore --backup ID --force 恢复原 DACL")
	}
	if err = validatePath(b.Path); err != nil {
		return err
	}
	if b.Recursive {
		// New links/hard links may have appeared since grant. Recheck and pin
		// the current tree before allowing Windows to propagate a restored ACL.
		count := 0
		err = filepath.WalkDir(b.Path, func(path string, _ fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			count++
			if count > maxObjects {
				return fmt.Errorf("恢复前检查超过 %d 个对象，请由管理员核对备份并缩小目录树", maxObjects)
			}
			_, err := m.open(path, false)
			return err
		})
		if err != nil {
			return fmt.Errorf("恢复前检查当前目录树失败，目标权限未修改：%w", err)
		}
	}
	// Check the entire backup before changing any object. --force only bypasses
	// ACL drift, never replacement identities, reparse points or missing paths.
	for i := range b.Objects {
		obj := &b.Objects[i]
		if err = validatePath(obj.Path); err != nil || !overlaps(b.Path, obj.Path) || (obj.Path != b.Path && !strings.HasPrefix(strings.ToLower(obj.Path), strings.ToLower(b.Path)+`\`)) {
			return errors.New("备份对象路径异常，未恢复")
		}
		current, err := m.open(obj.Path, true)
		if err != nil {
			return fmt.Errorf("恢复前检查失败，备份保留（请恢复被移动/删除的原对象后重试）：%w", err)
		}
		if current.Identity != obj.Identity || current.Directory != obj.Directory {
			return fmt.Errorf("对象已被替换，拒绝把旧权限应用到新对象：%s", obj.Path)
		}
		obj.handle = current.handle
		after, err := descriptor(*obj)
		if err != nil {
			return err
		}
		if !o.Force && comparableSDDL(after) != comparableSDDL(obj.After) {
			return fmt.Errorf("授权后 ACL/属主发生变化：%s；未恢复，请核对备份后用 --force 明确覆盖后续 ACL 修改", obj.Path)
		}
	}
	// Journal before restoring so a crash/partial failure remains recoverable.
	b.Status = "prepared"
	if err = m.save(*b); err != nil {
		return err
	}
	if err = m.publish(records); err != nil {
		return err
	}
	var failures error
	for _, obj := range b.Objects {
		failures = errors.Join(failures, setDACL(obj, obj.Before))
	}
	if failures != nil {
		return fmt.Errorf("恢复未完成，Git 信任已撤销，备份保留；排除错误后使用原 ID 加 --force 重试：%w", failures)
	}
	b.Status = "restored"
	if err = m.save(*b); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "已恢复备份 %s 的原 DACL 并撤销本次 Git 信任；属主/身份未修改，备份保留。\n", b.ID)
	return err
}
