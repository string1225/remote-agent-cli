package store

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestEndpointGuideConcurrentCopyAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "guide.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	for _, u := range []User{{ID: "admin", Username: "admin", Admin: true}, {ID: "owner", Username: "owner"}} {
		if err = db.CreateUser(u); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	secrets := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s, e := db.EnsureEndpointGuideToken("admin", "owner")
			if e != nil {
				t.Error(e)
			}
			secrets <- s
		}()
	}
	wg.Wait()
	close(secrets)
	var first string
	for s := range secrets {
		if s == "" {
			t.Fatal("empty token")
		}
		if first == "" {
			first = s
		}
		if s != first {
			t.Fatal("concurrent copy issued duplicates")
		}
	}
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_, restored, err := db.EnsureEndpointGuideToken("admin", "owner")
	if err != nil || restored != first {
		t.Fatal("guide not restored", err)
	}
	if _, s, e := db.EnsureEndpointGuideToken("owner", "owner"); e == nil || s != "" {
		t.Fatal("nonadmin issuance accepted")
	}
}
