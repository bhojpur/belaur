package adapter

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
	"io/ioutil"
	"os"
	"testing"

	plcsvc "github.com/bhojpur/policy/pkg/engine"
	"github.com/bhojpur/policy/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	bolt "go.etcd.io/bbolt"
)

const testDB = "test.db"

type AdapterTestSuite struct {
	suite.Suite
	db       *bolt.DB
	enforcer plcsvc.IEnforcer
}

func testGetPolicy(t *testing.T, e plcsvc.IEnforcer, wanted [][]string) {
	t.Helper()
	got := e.GetPolicy()
	if !util.Array2DEquals(wanted, got) {
		t.Error("got policy: ", got, ", wanted policy: ", wanted)
	}
}

func (suite *AdapterTestSuite) SetupTest() {
	t := suite.T()

	db, err := bolt.Open(testDB, 0600, nil)
	if err != nil {
		t.Fatalf("error opening bolt db: %s\n", err.Error())
	}
	suite.db = db

	bts, err := ioutil.ReadFile("examples/rbac_policy.csv")
	if err != nil {
		t.Error(err)
	}

	a, err := NewAdapter(db, "policy", string(bts))
	if err != nil {
		t.Error(err)
	}

	enforcer, err := plcsvc.NewEnforcer("examples/rbac_model.conf", a)
	if err != nil {
		t.Errorf("error creating enforcer: %s\n", err.Error())
	}

	suite.enforcer = enforcer
}

func (suite *AdapterTestSuite) TearDownTest() {
	suite.db.Close()
	if _, err := os.Stat(testDB); err == nil {
		os.Remove(testDB)
	}
}

func Test_AdapterTest_Suite(t *testing.T) {
	suite.Run(t, new(AdapterTestSuite))
}

func (suite *AdapterTestSuite) Test_LoadBuiltinPolicy() {
	testGetPolicy(suite.T(), suite.enforcer, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func (suite *AdapterTestSuite) Test_SavePolicy_ReturnsErr() {
	e := suite.enforcer
	t := suite.T()

	err := e.SavePolicy()
	assert.EqualError(t, err, "not supported: must use auto-save with this adapter")
}

func (suite *AdapterTestSuite) Test_AutoSavePolicy() {
	e := suite.enforcer
	t := suite.T()

	e.EnableAutoSave(true)

	e.AddPolicy("roger", "data1", "write")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"roger", "data1", "write"}})

	e.RemovePolicy("roger", "data1", "write")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	e.AddPolicies([][]string{{"roger", "data1", "read"}, {"roger", "data1", "write"}})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"roger", "data1", "read"}, {"roger", "data1", "write"}})

	e.RemoveFilteredPolicy(0, "roger")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	_, err := e.RemoveFilteredPolicy(1, "data1")
	assert.EqualError(t, err, "fieldIndex != 0: adapter only supports filter by prefix")

	e.AddPolicies([][]string{{"roger", "data1", "read"}, {"roger", "data1", "write"}})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"roger", "data1", "read"}, {"roger", "data1", "write"}})

	e.RemovePolicies([][]string{{"roger", "data1", "read"}, {"roger", "data1", "write"}})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}
