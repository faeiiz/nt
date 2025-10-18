package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var NotesBucket = []byte("notes")
var MetaBucket = []byte("meta")
var IDKey = []byte("id_seq")

type Note struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Completed bool      `json:"completed"`
}

type BoltStore struct {
	db *bolt.DB
}

func DefaultDBPath() (string, error) {
	dir := filepath.Join(os.Getenv("HOME"), ".local", "share", "nt")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "notes.db"), nil
}

func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(NotesBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(MetaBucket); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &BoltStore{db: db}, nil
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}

func (s *BoltStore) nextID(tx *bolt.Tx) (uint64, error) {
	b := tx.Bucket(MetaBucket)
	seq := b.Get(IDKey)
	var id uint64
	if seq == nil {
		id = 1
	} else {
		id = bytesToUint64(seq) + 1
	}
	if err := b.Put(IDKey, uint64ToBytes(id)); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *BoltStore) Add(title, body string) (*Note, error) {
	var n *Note
	err := s.db.Update(func(tx *bolt.Tx) error {
		nb := tx.Bucket(NotesBucket)
		id, err := s.nextID(tx)
		if err != nil {
			return err
		}
		n = &Note{
			ID:        id,
			Title:     title,
			Body:      body,
			CreatedAt: time.Now().UTC(),
		}
		js, err := json.Marshal(n)
		if err != nil {
			return err
		}
		return nb.Put(uint64ToBytes(id), js)
	})
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (s *BoltStore) List() ([]*Note, error) {
	var out []*Note
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(NotesBucket)
		return b.ForEach(func(k, v []byte) error {
			var n Note
			if err := json.Unmarshal(v, &n); err != nil {
				return err
			}
			out = append(out, &n)
			return nil
		})
	})
	return out, err
}

func (s *BoltStore) Update(n *Note) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(NotesBucket)
		js, err := json.Marshal(n)
		if err != nil {
			return err
		}
		return b.Put(uint64ToBytes(n.ID), js)
	})
}

func (s *BoltStore) Delete(id uint64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(NotesBucket)
		return b.Delete(uint64ToBytes(id))
	})
}

func (s *BoltStore) Get(id uint64) (*Note, error) {
	var n Note
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(NotesBucket)
		v := b.Get(uint64ToBytes(id))
		if v == nil {
			return errors.New("not found")
		}
		return json.Unmarshal(v, &n)
	})
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// helpers
func uint64ToBytes(u uint64) []byte {
	b := make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		b[7-i] = byte(u >> (i * 8))
	}
	return b
}
func bytesToUint64(b []byte) uint64 {
	var u uint64
	for i := uint(0); i < 8; i++ {
		u |= uint64(b[7-i]) << (i * 8)
	}
	return u
}
