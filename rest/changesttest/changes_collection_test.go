//  Copyright 2022-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package changestest

import (
	"log"
	"testing"

	"github.com/couchbase/sync_gateway/channels"

	"github.com/couchbase/sync_gateway/db"
	"github.com/stretchr/testify/assert"

	"github.com/couchbase/sync_gateway/base"
	"github.com/couchbase/sync_gateway/rest"
	"github.com/stretchr/testify/require"
)

func TestMultiCollectionChangesAdmin(t *testing.T) {

	base.SetUpTestLogging(t, base.LevelDebug, base.KeyHTTP, base.KeyChanges, base.KeyCache)
	testBucket := base.GetTestBucket(t)
	scopesConfig, keyspaces := getMultiCollectionConfig(t, testBucket, 2)
	rtConfig := &rest.RestTesterConfig{
		CustomTestBucket: testBucket,
		DatabaseConfig: &rest.DatabaseConfig{
			DbConfig: rest.DbConfig{
				Scopes: scopesConfig,
			},
		},
	}

	c1Keyspace := keyspaces[0]
	c2Keyspace := keyspaces[1]

	rt := rest.NewRestTester(t, rtConfig)
	defer rt.Close()

	// Put several documents, will be retrieved via query
	response := rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs1", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/abc1", `{"value":1, "channels":["ABC"]}`)
	rest.RequireStatus(t, response, 201)

	_ = rt.WaitForPendingChanges()

	var changes struct {
		Results  []db.ChangeEntry
		Last_Seq db.SequenceID
	}

	// Issue changes request.  Will initialize cache for channels, and return docs via query
	changesResponse := rt.SendAdminRequest("GET", "/"+c1Keyspace+"/_changes?since=0", "")
	err := base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendAdminRequest("GET", "/"+c2Keyspace+"/_changes?since=0", "")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	// Put more documents, should be served via DCP/cache
	response = rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs2", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/abc2", `{"value":1, "channels":["ABC"]}`)
	rest.RequireStatus(t, response, 201)
	_ = rt.WaitForPendingChanges()

	changesResponse = rt.SendAdminRequest("GET", "/"+c1Keyspace+"/_changes?since=0", "")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 2)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendAdminRequest("GET", "/"+c2Keyspace+"/_changes?since=0", "")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 2)
	logChangesResponse(t, changesResponse.Body.Bytes())
}

func TestMultiCollectionChangesAdminSameChannelName(t *testing.T) {

	base.SetUpTestLogging(t, base.LevelDebug, base.KeyHTTP, base.KeyChanges, base.KeyCache)
	testBucket := base.GetTestBucket(t)
	scopesConfig, keyspaces := getMultiCollectionConfig(t, testBucket, 2)
	rtConfig := &rest.RestTesterConfig{
		CustomTestBucket: testBucket,
		DatabaseConfig: &rest.DatabaseConfig{
			DbConfig: rest.DbConfig{
				Scopes: scopesConfig,
			},
		},
	}

	c1Keyspace := keyspaces[0]
	c2Keyspace := keyspaces[1]

	rt := rest.NewRestTester(t, rtConfig)
	defer rt.Close()

	// Put several documents, will be retrieved via query
	response := rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs1_c1", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/pbs1_c2", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	_ = rt.WaitForPendingChanges()

	var changes struct {
		Results  []db.ChangeEntry
		Last_Seq db.SequenceID
	}

	// Issue changes request.  Will initialize cache for channels, and return docs via query
	changesResponse := rt.SendAdminRequest("GET", "/"+c1Keyspace+"/_changes?since=0", "")
	err := base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendAdminRequest("GET", "/"+c2Keyspace+"/_changes?since=0", "")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	// Put more documents, should be served via DCP/cache
	response = rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs2_c1", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/pbs2_c2", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	_ = rt.WaitForPendingChanges()

	changesResponse = rt.SendAdminRequest("GET", "/"+c1Keyspace+"/_changes?since=0", "")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 2)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendAdminRequest("GET", "/"+c2Keyspace+"/_changes?since=0", "")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 2)
	logChangesResponse(t, changesResponse.Body.Bytes())
}

func TestMultiCollectionChangesUser(t *testing.T) {

	base.SetUpTestLogging(t, base.LevelDebug, base.KeyHTTP, base.KeyChanges, base.KeyCache, base.KeyCRUD)
	testBucket := base.GetTestBucket(t)
	scopesConfig, keyspaces := getMultiCollectionConfig(t, testBucket, 2)
	rtConfig := &rest.RestTesterConfig{
		CustomTestBucket: testBucket,
		DatabaseConfig: &rest.DatabaseConfig{
			DbConfig: rest.DbConfig{
				Scopes: scopesConfig,
			},
		},
	}

	c1Keyspace := keyspaces[0]
	c2Keyspace := keyspaces[1]

	rt := rest.NewRestTester(t, rtConfig)
	defer rt.Close()

	// Create user with access to channel PBS in both collections
	ctx := rt.Context()
	a := rt.ServerContext().Database(ctx, "db").Authenticator(ctx)
	bernard, err := a.NewUser("bernard", "letmein", channels.BaseSetOf(t, "PBS"))
	assert.NoError(t, err)
	assert.NoError(t, a.Save(bernard))

	// Put several documents, will be retrieved via query
	response := rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs1_c1", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/pbs1_c2", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	_ = rt.WaitForPendingChanges()

	var changes struct {
		Results  []db.ChangeEntry
		Last_Seq db.SequenceID
	}

	// Issue changes request.  Will initialize cache for channels, and return docs via query
	changesResponse := rt.SendUserRequest("GET", "/"+c1Keyspace+"/_changes?since=0", "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendUserRequest("GET", "/"+c2Keyspace+"/_changes?since=0", "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	// Put more documents, should be served via DCP/cache
	response = rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs2_c1", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/pbs2_c2", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	_ = rt.WaitForPendingChanges()

	changesResponse = rt.SendUserRequest("GET", "/"+c1Keyspace+"/_changes?since=0", "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 2)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendUserRequest("GET", "/"+c2Keyspace+"/_changes?since=0", "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 2)
	logChangesResponse(t, changesResponse.Body.Bytes())
}

// TestMultiCollectionChangesUserDynamicGrant tests a dynamic channel grant when that channel is not already resident
// in the cache (post-grant changes triggers query backfill of the cache)
func TestMultiCollectionChangesUserDynamicGrant(t *testing.T) {

	base.SetUpTestLogging(t, base.LevelDebug, base.KeyHTTP, base.KeyChanges, base.KeyCache, base.KeyCRUD)
	testBucket := base.GetTestBucket(t)
	scopesConfig, keyspaces := getMultiCollectionConfig(t, testBucket, 2)
	rtConfig := &rest.RestTesterConfig{
		CustomTestBucket: testBucket,
		DatabaseConfig: &rest.DatabaseConfig{
			DbConfig: rest.DbConfig{
				Scopes: scopesConfig,
			},
		},
	}

	c1Keyspace := keyspaces[0]
	c2Keyspace := keyspaces[1]

	rt := rest.NewRestTester(t, rtConfig)
	defer rt.Close()

	// Create user with access to channel PBS in both collections
	ctx := rt.Context()
	a := rt.ServerContext().Database(ctx, "db").Authenticator(ctx)
	bernard, err := a.NewUser("bernard", "letmein", channels.BaseSetOf(t, "PBS"))
	assert.NoError(t, err)
	assert.NoError(t, a.Save(bernard))

	// Put several documents
	response := rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs1_c1", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/abc1_c1", `{"value":1, "channels":["ABC"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/pbs1_c2", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/abc1_c2", `{"value":1, "channels":["ABC"]}`)
	rest.RequireStatus(t, response, 201)
	_ = rt.WaitForPendingChanges()

	var changes struct {
		Results  []db.ChangeEntry
		Last_Seq db.SequenceID
	}

	// Issue changes request.  Will initialize cache for channels, and return docs via query
	changesResponse := rt.SendUserRequest("GET", "/"+c1Keyspace+"/_changes?since=0", "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendUserRequest("GET", "/"+c2Keyspace+"/_changes?since=0", "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())
	lastSeq := changes.Last_Seq

	// Grant user access to channel ABC in collection 1
	err = rt.SetAdminChannels("bernard", c1Keyspace, "ABC", "PBS")
	require.NoError(t, err)

	// confirm that change from c1 is sent, along with user doc
	changesResponse = rt.SendUserRequest("GET", "/"+c1Keyspace+"/_changes?since="+lastSeq.String(), "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 2)
	logChangesResponse(t, changesResponse.Body.Bytes())

	// Confirm that access hasn't been granted in c2, expect only user doc
	changesResponse = rt.SendUserRequest("GET", "/"+c2Keyspace+"/_changes?since="+lastSeq.String(), "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())
}

// TestMultiCollectionChangesUserDynamicGrant tests a dynamic channel grant when that channel is resident
// in the cache (verifies cache buffering of principals)
func TestMultiCollectionChangesUserDynamicGrantDCP(t *testing.T) {

	base.SetUpTestLogging(t, base.LevelDebug, base.KeyHTTP, base.KeyChanges, base.KeyCache, base.KeyCRUD)
	testBucket := base.GetTestBucket(t)
	scopesConfig, keyspaces := getMultiCollectionConfig(t, testBucket, 2)
	rtConfig := &rest.RestTesterConfig{
		CustomTestBucket: testBucket,
		DatabaseConfig: &rest.DatabaseConfig{
			DbConfig: rest.DbConfig{
				Scopes: scopesConfig,
			},
		},
	}

	c1Keyspace := keyspaces[0]
	c2Keyspace := keyspaces[1]

	rt := rest.NewRestTester(t, rtConfig)
	defer rt.Close()

	// Create user with access to channel PBS in both collections
	ctx := rt.Context()
	a := rt.ServerContext().Database(ctx, "db").Authenticator(ctx)
	bernard, err := a.NewUser("bernard", "letmein", channels.BaseSetOf(t, "PBS"))
	assert.NoError(t, err)
	assert.NoError(t, a.Save(bernard))

	// Put several documents
	response := rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs1_c1", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/abc1_c1", `{"value":1, "channels":["ABC"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/pbs1_c2", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/abc1_c2", `{"value":1, "channels":["ABC"]}`)
	rest.RequireStatus(t, response, 201)
	_ = rt.WaitForPendingChanges()

	var changes struct {
		Results  []db.ChangeEntry
		Last_Seq db.SequenceID
	}

	// Issue changes request.  Will initialize cache for user channel (PBS), and return docs via query
	changesResponse := rt.SendUserRequest("GET", "/"+c1Keyspace+"/_changes?since=0", "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendUserRequest("GET", "/"+c2Keyspace+"/_changes?since=0", "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())
	lastSeq := changes.Last_Seq

	// Issue admin changes request for channel ABC, to initialize cache for that channel in each collection
	changesResponse = rt.SendAdminRequest("GET", "/"+c1Keyspace+"/_changes?filter=sync_gateway/bychannel&channels=ABC", "")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	changesResponse = rt.SendAdminRequest("GET", "/"+c2Keyspace+"/_changes?filter=sync_gateway/bychannel&channels=ABC", "")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	require.Len(t, changes.Results, 1)
	logChangesResponse(t, changesResponse.Body.Bytes())

	// Grant user access to channel ABC in collection 1
	err = rt.SetAdminChannels("bernard", c1Keyspace, "ABC", "PBS")
	require.NoError(t, err)

	// Write additional docs to the cached channels, should be served via DCP/cache
	response = rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/abc2_c1", `{"value":1, "channels":["ABC"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/abc2_c2", `{"value":1, "channels":["ABC"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c1Keyspace+"/pbs2_c1", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	response = rt.SendAdminRequest("PUT", "/"+c2Keyspace+"/pbs2_c2", `{"value":1, "channels":["PBS"]}`)
	rest.RequireStatus(t, response, 201)
	_ = rt.WaitForPendingChanges()

	// Expect 5 documents in collection with ABC grant:
	//  - backfill of 1 ABC doc written prior to lastSeq
	//  - user doc
	//  - 1 PBS docs after lastSeq
	//  - 1 ABC docs after lastSeq
	changesResponse = rt.SendUserRequest("GET", "/"+c1Keyspace+"/_changes?since="+lastSeq.String(), "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	assert.Len(t, changes.Results, 4)
	logChangesResponse(t, changesResponse.Body.Bytes())

	// Expect 2 documents in collection without ABC grant
	//  - user doc
	//  - 1 PBS doc after lastSeq
	changesResponse = rt.SendUserRequest("GET", "/"+c2Keyspace+"/_changes?since="+lastSeq.String(), "", "bernard")
	err = base.JSONUnmarshal(changesResponse.Body.Bytes(), &changes)
	assert.NoError(t, err, "Error unmarshalling changes response")
	assert.Len(t, changes.Results, 2)
	logChangesResponse(t, changesResponse.Body.Bytes())
}

func logChangesResponse(t *testing.T, response []byte) {
	var changes struct {
		Results  []db.ChangeEntry
		Last_Seq db.SequenceID
	}

	err := base.JSONUnmarshal(response, &changes)
	require.NoError(t, err, "unable to marshal changes for logging")

	for _, entry := range changes.Results {
		log.Printf("changes Response entry: %+v", entry)
	}
	return
}

// getScopesConfig sets up a ScopesConfig from a TestBucket and uses a non default collection if available.
func getMultiCollectionConfig(t testing.TB, testBucket *base.TestBucket, numCollections int) (rest.ScopesConfig, []string) {
	// Get a datastore as provided by the test
	stores, err := testBucket.ListDataStores()
	require.NoError(t, err)
	require.True(t, len(stores) >= numCollections, "Requested more collections than found on testBucket")

	scopesConfig := rest.ScopesConfig{}
	keyspaces := make([]string, numCollections)
	for i := 0; i < numCollections; i++ {
		dataStoreName := stores[i]
		keyspaces[i] = "db." + dataStoreName.ScopeName() + "." + dataStoreName.CollectionName()
		if scopeConfig, ok := scopesConfig[dataStoreName.ScopeName()]; ok {
			if _, ok := scopeConfig.Collections[dataStoreName.CollectionName()]; ok {
				// already present
			} else {
				scopeConfig.Collections[dataStoreName.CollectionName()] = rest.CollectionConfig{}
			}
		} else {
			scopesConfig[dataStoreName.ScopeName()] = rest.ScopeConfig{
				Collections: map[string]rest.CollectionConfig{
					dataStoreName.CollectionName(): {},
				}}
		}

	}
	return scopesConfig, keyspaces
}