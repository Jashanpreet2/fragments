package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFragmentDBInterface(t *testing.T) {
	teardown := PreTestSetup()
	defer teardown()

	t.Run("TestWriteFragment", func(t *testing.T) {
		frag := Fragment{"1", "user", time.Now(), time.Now(), "testType", 1}
		assert.Equal(t, true, WriteFragment(&frag))
	})

	t.Run("TestReadFragmentSuccess", func(t *testing.T) {
		userId := "user"
		fragmentId := "1"
		frag := Fragment{fragmentId, userId, time.Now(), time.Now(), "testType", 1}
		ok := assert.Equal(t, true, WriteFragment(&frag))
		if !ok {
			t.Error("Failed to create fragment")
			return
		}
		_, ok = ReadFragment(userId, fragmentId)
		assert.Equal(t, true, ok)
	})

	t.Run("TestReadFragmentFail", func(t *testing.T) {
		_, ok := ReadFragment("Non existent User ID", "Non existent Fragment ID")
		assert.Equal(t, false, ok)
	})

	t.Run("TestWriteFragmentData", func(t *testing.T) {
		userId := "user"
		fragmentId := "1"
		ok := WriteFragmentData(userId, fragmentId, []byte("Sample data"))
		assert.Equal(t, true, ok)
	})

	t.Run("TestReadFragmentData", func(t *testing.T) {
		userId := "user"
		fragmentId := "1"
		data := []byte("Sample data")
		if ok := WriteFragmentData(userId, fragmentId, data); !ok {
			t.Error("Failed to write user fragment")
		}
		retrievedData, ok := ReadFragmentData(userId, fragmentId)
		assert.Equal(t, ok, true)
		assert.Equal(t, data, retrievedData)
	})

	t.Run("TestDeleteFragmentDb", func(t *testing.T) {
		frag := CreateTestFragment()
		data := []byte("Sample data")
		ok := WriteFragmentData(frag.OwnerId, frag.Id, data)
		if !ok {
			t.Error("Failed to write fragment data")
			return
		}
		ok = DeleteFragmentDB(frag.OwnerId, frag.Id)
		assert.Equal(t, true, ok)
	})

	t.Run("TestListFragmentIds", func(t *testing.T) {
		frag := CreateTestFragment()
		ok := WriteFragment(&frag)
		if !ok {
			t.Error("Failed to write fragment")
		}
		ids := GetUserFragmentIds(frag.OwnerId)
		assert.Equal(t, []string{frag.Id}, ids)
	})

	t.Run("TestListFragmentMetadatas", func(t *testing.T) {
		frag := CreateTestFragment()
		ok := WriteFragment(&frag)
		if !ok {
			t.Error("Failed to write fragment")
		}
		retrievedFrags := ListFragmentMetadatas(frag.OwnerId)
		assert.Equal(t, []Fragment{frag}, retrievedFrags)
	})

	t.Run("TestGenerateID", func(t *testing.T) {
		assert.Equal(t, 0, GenerateID("Non-existent user"))
	})

	t.Run("TestReadFragmentDataFails", func(t *testing.T) {
		_, ok := ReadFragmentData("NonExistentUser", "NonExistentId")
		assert.Equal(t, false, ok)
	})

	t.Run("TestDeleteFragmentDbDataDoesn'tExist", func(t *testing.T) {
		ResetDB()
		fragment := CreateTestFragment()
		WriteFragment(&fragment)
		ok := DeleteFragmentDB(fragment.OwnerId, fragment.Id)
		assert.Equal(t, false, ok)
	})

	t.Run("TestDeleteFragmentDbFragmentDoesn'tExist", func(t *testing.T) {
		ResetDB()
		fragment := CreateTestFragment()
		WriteFragmentData(fragment.OwnerId, fragment.Id, []byte("Some data"))
		ok := DeleteFragmentDB(fragment.OwnerId, fragment.Id)
		assert.Equal(t, false, ok)
	})
}
