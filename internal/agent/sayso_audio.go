package agent

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/store"
)

const soAudioLimit = 64 << 20

type soUpload struct {
	owner     *peer
	sessionID string
	path      string
	meta      soRecording
	written   int64
	updated   time.Time
}

func (s *saysoStore) audio(p *peer, method string, parts []string, b map[string]json.RawMessage) (any, error) {
	if len(parts) == 2 && parts[1] == "begin" && method == "POST" {
		for id, u := range s.uploads {
			if u.owner.ctx.Err() != nil || time.Since(u.updated) > 5*time.Minute {
				_ = os.Remove(u.path)
				delete(s.uploads, id)
			}
		}
		for _, u := range s.uploads {
			if u.owner == p {
				return nil, errors.New("请等待当前录音上传完成")
			}
		}
		if len(s.uploads) >= 8 {
			return nil, errors.New("录音上传繁忙")
		}
		id, e := soString(b, "sessionId", 100, true)
		if e != nil {
			return nil, e
		}
		v := s.session(id)
		if v == nil {
			return nil, errors.New("会话不存在")
		}
		if len(v.Recordings) >= 1000 {
			return nil, errors.New("单会话最多 1000 段录音")
		}
		var meta soRecording
		raw, _ := json.Marshal(b)
		if json.Unmarshal(raw, &meta) != nil {
			return nil, errors.New("无效录音信息")
		}
		if meta.SizeBytes < 1 || meta.SizeBytes > soAudioLimit || meta.DurationMS < 0 || meta.DurationMS > 11*60*1000 {
			return nil, errors.New("录音分段大小或时长超限")
		}
		mime := strings.ToLower(strings.Split(meta.MimeType, ";")[0])
		ext := ""
		switch mime {
		case "audio/webm", "video/webm":
			ext = ".webm"
		case "audio/ogg":
			ext = ".ogg"
		case "audio/mp4":
			ext = ".m4a"
		case "audio/wav", "audio/x-wav":
			ext = ".wav"
		default:
			return nil, errors.New("不支持的录音格式")
		}
		if len(meta.MimeType) > 100 {
			return nil, errors.New("无效录音格式")
		}
		start, e := time.Parse(time.RFC3339Nano, meta.StartedAt)
		if e != nil {
			return nil, errors.New("无效录音开始时间")
		}
		end, e := time.Parse(time.RFC3339Nano, meta.EndedAt)
		if e != nil || end.Before(start) {
			return nil, errors.New("无效录音结束时间")
		}
		meta.ID = store.Token()
		meta.Filename = meta.ID + ext
		f, e := os.CreateTemp(filepath.Join(s.dir, "recordings"), ".upload-*")
		if e != nil {
			return nil, e
		}
		path := f.Name()
		_ = f.Close()
		u := &soUpload{p, id, path, meta, 0, time.Now()}
		s.uploads[meta.ID] = u
		go func() {
			<-p.ctx.Done()
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.uploads[meta.ID] == u {
				_ = os.Remove(u.path)
				delete(s.uploads, meta.ID)
			}
		}()
		return map[string]string{"id": meta.ID}, nil
	}
	if len(parts) != 3 {
		return nil, errors.New("不支持的录音操作")
	}
	if parts[2] == "read" && method == "POST" {
		var meta *soRecording
		for _, v := range s.state.Sessions {
			for i := range v.Recordings {
				if v.Recordings[i].ID == parts[1] {
					meta = &v.Recordings[i]
				}
			}
		}
		if meta == nil {
			return nil, errors.New("录音不存在")
		}
		var offset int64
		if json.Unmarshal(b["offset"], &offset) != nil || offset < 0 || offset >= meta.SizeBytes {
			return nil, errors.New("无效读取位置")
		}
		f, e := os.Open(filepath.Join(s.dir, "recordings", filepath.Base(meta.Filename)))
		if e != nil {
			return nil, e
		}
		defer f.Close()
		data := make([]byte, min(int64(16000), meta.SizeBytes-offset))
		n, e := f.ReadAt(data, offset)
		if e != nil && e != io.EOF {
			return nil, e
		}
		if n != len(data) {
			return nil, errors.New("录音文件不完整")
		}
		return map[string]any{"data": data, "size": meta.SizeBytes, "mimeType": meta.MimeType}, nil
	}
	u := s.uploads[parts[1]]
	if u == nil || u.owner != p {
		return nil, errors.New("上传不存在或已失效")
	}
	if method != "POST" {
		return nil, errors.New("不支持的录音操作")
	}
	switch parts[2] {
	case "abort":
		_ = os.Remove(u.path)
		delete(s.uploads, parts[1])
		return true, nil
	case "chunk":
		var offset int64
		if json.Unmarshal(b["offset"], &offset) != nil || offset != u.written {
			return nil, errors.New("录音片段顺序不正确")
		}
		encoded, e := soString(b, "data", 24000, true)
		if e != nil {
			return nil, e
		}
		data, e := base64.StdEncoding.DecodeString(encoded)
		if e != nil || len(data) == 0 || len(data) > 16000 || u.written+int64(len(data)) > u.meta.SizeBytes {
			return nil, errors.New("无效录音分片")
		}
		f, e := os.OpenFile(u.path, os.O_WRONLY|os.O_APPEND, 0600)
		if e != nil {
			return nil, e
		}
		n, e := f.Write(data)
		closeErr := f.Close()
		if e == nil {
			e = closeErr
		}
		if e != nil {
			return nil, e
		}
		u.written += int64(n)
		u.updated = time.Now()
		return map[string]int64{"offset": u.written}, nil
	case "finish":
		if u.written != u.meta.SizeBytes {
			return nil, errors.New("录音尚未传完")
		}
		v := s.session(u.sessionID)
		if v == nil {
			return nil, errors.New("会话不存在")
		}
		if len(v.Recordings) >= 1000 {
			return nil, errors.New("录音数量超限")
		}
		f, e := os.OpenFile(u.path, os.O_RDWR, 0600)
		if e != nil {
			return nil, e
		}
		e = f.Sync()
		_ = f.Close()
		if e != nil {
			return nil, e
		}
		if e = os.Rename(u.path, filepath.Join(s.dir, "recordings", u.meta.Filename)); e != nil {
			return nil, e
		}
		v.Recordings = append(v.Recordings, u.meta)
		delete(s.uploads, parts[1])
		if e = s.save(); e != nil {
			return nil, e
		}
		return u.meta, nil
	default:
		return nil, errors.New("不支持的上传操作")
	}
}
