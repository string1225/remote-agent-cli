package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

var sharedFileID = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
var sharedFileBucket = regexp.MustCompile(`^[0-9]{10,12}$`)
var sharingTimezone = time.FixedZone("Asia/Shanghai", 8*60*60)

func nextFileCleanup(now time.Time) time.Time {
	local := now.In(sharingTimezone)
	cutoff := time.Date(local.Year(), local.Month(), local.Day(), 4, 0, 0, 0, sharingTimezone)
	if !now.Before(cutoff) {
		cutoff = cutoff.AddDate(0, 0, 1)
	}
	return cutoff
}

func fileBucket(expires time.Time) string { return strconv.FormatInt(expires.Unix(), 10) }

func (s *Server) uploadSharedFile(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateAgent(r); err != nil {
		fail(w, 401, "device credentials denied")
		return
	}
	if s.Config.SharedFiles == "" {
		fail(w, 503, "file sharing is not configured")
		return
	}
	if r.Header.Get("Content-Type") != "application/octet-stream" {
		fail(w, 415, "application/octet-stream required")
		return
	}
	name, err := url.QueryUnescape(r.Header.Get("X-File-Name"))
	if err != nil || name == "" || name == "." || name == ".." || len(name) > 240 || !utf8.ValidString(name) || strings.ContainsAny(name, `/\`) || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		fail(w, 400, "invalid file name")
		return
	}
	if r.ContentLength > protocol.MaxSharedFileBytes {
		fail(w, 413, "file exceeds 100 MiB")
		return
	}
	select {
	case s.sharedFileSlots <- struct{}{}:
		defer func() { <-s.sharedFileSlots }()
	default:
		fail(w, 429, "too many concurrent uploads; retry later")
		return
	}
	_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(15 * time.Minute))
	r.Body = http.MaxBytesReader(w, r.Body, protocol.MaxSharedFileBytes)
	expires := nextFileCleanup(time.Now())
	root, err := os.OpenRoot(s.Config.SharedFiles)
	if err != nil {
		fail(w, 503, "file storage unavailable")
		return
	}
	defer root.Close()
	bucket := fileBucket(expires)
	if err := root.Mkdir(bucket, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		fail(w, 500, "cannot create upload storage")
		return
	}
	file := protocol.SharedFile{ID: store.Token(), Name: name, ExpiresAt: expires}
	path := bucket + "/" + file.ID
	if err := root.Mkdir(path, 0700); err != nil {
		fail(w, 500, "cannot create upload")
		return
	}
	complete := false
	defer func() {
		if !complete {
			_ = root.RemoveAll(path)
		}
	}()
	out, err := root.OpenFile(path+"/data", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		fail(w, 500, "cannot save upload")
		return
	}
	hash := sha256.New()
	file.Size, err = io.Copy(io.MultiWriter(out, hash), r.Body)
	if err == nil {
		err = out.Sync()
	}
	closeErr := out.Close()
	if err != nil || closeErr != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			fail(w, 413, "file exceeds 100 MiB")
		} else {
			fail(w, 400, "upload interrupted; retry the file")
		}
		return
	}
	if !time.Now().Before(expires) {
		fail(w, 409, "daily cleanup occurred during upload; retry the file")
		return
	}
	file.SHA256 = hex.EncodeToString(hash.Sum(nil))
	metadata, _ := json.Marshal(file)
	if err := root.WriteFile(path+"/metadata.json", metadata, 0600); err != nil {
		fail(w, 500, "cannot publish upload")
		return
	}
	complete = true
	file.URL = s.Config.PublicURL + "/files/" + file.ID
	if isHTMLReport(file.Name) {
		file.PreviewURL = file.URL + "?view=1"
	}
	reply(w, 200, file)
}

func isHTMLReport(name string) bool {
	name = strings.ToLower(name)
	return strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm")
}

func (s *Server) downloadSharedFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	id := r.PathValue("id")
	if s.Config.SharedFiles == "" || !sharedFileID.MatchString(id) {
		http.NotFound(w, r)
		return
	}
	root, err := os.OpenRoot(s.Config.SharedFiles)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer root.Close()
	path := fileBucket(nextFileCleanup(time.Now())) + "/" + id
	raw, err := root.ReadFile(path + "/metadata.json")
	var metadata protocol.SharedFile
	if err != nil || json.Unmarshal(raw, &metadata) != nil || metadata.ID != id || !time.Now().Before(metadata.ExpiresAt) {
		http.NotFound(w, r)
		return
	}
	file, err := root.Open(path + "/data")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() != metadata.Size {
		http.NotFound(w, r)
		return
	}
	disposition := "attachment"
	w.Header().Set("Content-Type", "application/octet-stream")
	if r.URL.Query().Get("view") == "1" {
		if !isHTMLReport(metadata.Name) {
			fail(w, http.StatusUnsupportedMediaType, "only HTML reports support preview")
			return
		}
		disposition = "inline"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Reports share the application host. Give them an opaque origin and no
		// scripts, forms, network access or remote assets; self-contained CSS is
		// sufficient for readable reports without access to account sessions.
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'; sandbox")
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": metadata.Name}))
	w.Header().Set("X-File-Expires-At", metadata.ExpiresAt.Format(time.RFC3339))
	http.ServeContent(w, r, metadata.Name, info.ModTime(), file)
}

// PruneSharedFiles removes only expired, application-owned day buckets. It also
// catches up after downtime without deleting files uploaded after the cutoff.
func (s *Server) PruneSharedFiles(now time.Time) error {
	if s.Config.SharedFiles == "" {
		return nil
	}
	root, err := os.OpenRoot(s.Config.SharedFiles)
	if err != nil {
		return err
	}
	defer root.Close()
	dir, err := root.Open(".")
	if err != nil {
		return err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if !sharedFileBucket.MatchString(name) {
			continue
		}
		cutoff, _ := strconv.ParseInt(name, 10, 64)
		if cutoff <= now.Unix() {
			if err := root.RemoveAll(name); err != nil {
				return fmt.Errorf("remove expired shared files: %w", err)
			}
		}
	}
	return nil
}

func (s *Server) RunFileCleanup(ctx context.Context) {
	for {
		next := nextFileCleanup(time.Now())
		if err := s.PruneSharedFiles(time.Now()); err != nil {
			log.Printf("shared file cleanup failed: %v", err)
			// Retry transient filesystem failures instead of waiting another day.
			if retry := time.Now().Add(time.Minute); retry.Before(next) {
				next = retry
			}
		}
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
