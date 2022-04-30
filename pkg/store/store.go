package store

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"time"

	boltadapter "github.com/bhojpur/belaur/pkg/adapter"
	"github.com/bhojpur/policy/pkg/persist"
	bolt "go.etcd.io/bbolt"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/assethelper"
	"github.com/bhojpur/belaur/pkg/security"
)

var (
	// Name of the bucket where we store user objects
	userBucket = []byte("Users")

	// Where we store all users permissions
	userPermsBucket = []byte("UserPermissions")

	// Name of the bucket where we store information about pipelines
	pipelineBucket = []byte("Pipelines")

	// Name of the bucket where we store information about pipelines
	// which are not yet compiled (create pipeline)
	createPipelineBucket = []byte("CreatePipelines")

	// Name of the bucket where we store all pipeline runs.
	pipelineRunBucket = []byte("PipelineRun")

	// Name of the bucket where we store information about settings
	settingsBucket = []byte("Settings")

	// Name of the bucket where we store all worker.
	workerBucket = []byte("Worker")

	// SHA pair bucket.
	shaPairBucket = []byte("SHAPair")
)

const (
	// Username and password of the first admin user
	adminUsername = "admin"
	adminPassword = "admin"
	autoUsername  = "auto"
	autoPassword  = "auto"

	// Bolt database file name
	boltDBFileName = "belaur.db"
)

// BoltStore represents the access type for store
type BoltStore struct {
	db            *bolt.DB
	policyAdapter persist.BatchAdapter
}

// SettingsStore is the interface that defines methods needed to save settings config into the store.
type SettingsStore interface {
	SettingsPut(config *belaur.StoreConfig) error
	SettingsGet() (*belaur.StoreConfig, error)
}

// BelaurStore is the interface that defines methods needed to store
// pipeline and user related information.
type BelaurStore interface {
	SettingsStore
	Init(dataPath string) error
	Close() error
	CreatePipelinePut(createPipeline *belaur.CreatePipeline) error
	CreatePipelineGet() (listOfPipelines []belaur.CreatePipeline, err error)
	PipelinePut(pipeline *belaur.Pipeline) error
	PipelineGet(id int) (pipeline *belaur.Pipeline, err error)
	PipelineGetByName(name string) (pipline *belaur.Pipeline, err error)
	PipelineGetRunHighestID(pipeline *belaur.Pipeline) (id int, err error)
	PipelinePutRun(r *belaur.PipelineRun) error
	PipelineGetScheduled(limit int) ([]*belaur.PipelineRun, error)
	PipelineGetRunByPipelineIDAndID(pipelineid int, runid int) (*belaur.PipelineRun, error)
	PipelineGetAllRuns() ([]belaur.PipelineRun, error)
	PipelineGetAllRunsByPipelineID(pipelineID int) ([]belaur.PipelineRun, error)
	PipelineGetLatestRun(pipelineID int) (*belaur.PipelineRun, error)
	PipelineGetRunByID(runID string) (*belaur.PipelineRun, error)
	PipelineDelete(id int) error
	PipelineRunDelete(uniqueID string) error
	UserPut(u *belaur.User, encryptPassword bool) error
	UserAuth(u *belaur.User, updateLastLogin bool) (*belaur.User, error)
	UserGet(username string) (*belaur.User, error)
	UserGetAll() ([]belaur.User, error)
	UserDelete(u string) error
	UserPermissionsPut(perms *belaur.UserPermission) error
	UserPermissionsGet(username string) (*belaur.UserPermission, error)
	UserPermissionsDelete(username string) error
	WorkerPut(w *belaur.Worker) error
	WorkerGetAll() ([]*belaur.Worker, error)
	WorkerDelete(id string) error
	WorkerDeleteAll() error
	WorkerGet(id string) (*belaur.Worker, error)
	UpsertSHAPair(pair belaur.SHAPair) error
	GetSHAPair(pipelineID int) (bool, belaur.SHAPair, error)
	BhojpurStore() persist.BatchAdapter
}

// Compile time interface compliance check for BoltStore. If BoltStore
// wouldn't implement BelaurStore this line wouldn't compile.
var _ BelaurStore = (*BoltStore)(nil)

// NewBoltStore creates a new instance of Store.
func NewBoltStore() *BoltStore {
	s := &BoltStore{}

	return s
}

// Init creates the data folder if not exists,
// generates private key and bolt database.
// This should be called only once per database
// because bolt holds a lock on the database file.
func (s *BoltStore) Init(dataPath string) error {
	// Open connection to bolt database
	path := filepath.Join(dataPath, boltDBFileName)
	// Give boltdb 5 seconds to try and open up a db file.
	// If another process is already holding that file, this will time-out.
	db, err := bolt.Open(path, belaur.Cfg.Bolt.Mode, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return err
	}
	s.db = db

	// TODO: Having Bhojpur Policy stuff here doesn't sit quite right with me (especially loading a file here).
	// Unfortunately we need to re-use the open bolt database for the adapter though.
	builtinPolicy, _ := assethelper.LoadRBACBuiltinPolicy()
	policyAdapter, err := boltadapter.NewAdapter(db, "bhojpur-policies", builtinPolicy)
	if err != nil {
		return err
	}
	s.policyAdapter = policyAdapter

	// Setup database
	return s.setupDatabase()
}

// Close closes the active boltdb connection.
func (s *BoltStore) Close() error {
	return s.db.Close()
}

type setup struct {
	bs  *BoltStore
	err error
}

// Create bucket if not exists function
func (s *setup) update(bucketName []byte) {
	// update is a no-op in case there was already an error
	if s.err != nil {
		return
	}
	c := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}
	s.err = s.bs.db.Update(c)
}

// setupDatabase create all buckets in the db.
// Additionally, it makes sure that the admin user exists.
func (s *BoltStore) setupDatabase() error {
	// Create bucket if not exists function
	setP := &setup{
		bs:  s,
		err: nil,
	}

	// Make sure buckets exist
	setP.update(userBucket)
	setP.update(userPermsBucket)
	setP.update(pipelineBucket)
	setP.update(createPipelineBucket)
	setP.update(pipelineRunBucket)
	setP.update(settingsBucket)
	setP.update(workerBucket)
	setP.update(shaPairBucket)

	if setP.err != nil {
		return setP.err
	}

	// Make sure that the user "admin" does exist
	admin, err := s.UserGet(adminUsername)
	if err != nil {
		return err
	}

	// Create admin user if we cannot find it
	if admin == nil {
		err = s.UserPut(&belaur.User{
			DisplayName: adminUsername,
			Username:    adminUsername,
			Password:    adminPassword,
		}, true)

		if err != nil {
			return err
		}
	}

	err = s.CreatePermissionsIfNotExisting()
	if err != nil {
		return err
	}

	u, err := s.UserGet(autoUsername)
	if err != nil {
		return err
	}

	if u == nil {
		triggerToken := security.GenerateRandomUUIDV5()
		auto := belaur.User{
			DisplayName:  "Auto User",
			JwtExpiry:    0,
			Password:     autoPassword,
			Tokenstring:  "",
			TriggerToken: triggerToken,
			Username:     autoUsername,
			LastLogin:    time.Now(),
		}
		err = s.UserPut(&auto, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// BhojpurStore is as a getter for the Bhojpur Policy store adapter.
func (s *BoltStore) BhojpurStore() persist.BatchAdapter {
	return s.policyAdapter
}
