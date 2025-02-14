package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFragmentDBInterface(t *testing.T) {
	teardown := PreTestSetup()
	defer teardown()

	t.Run("TestWriteFragment", func(t *testing.T) {
		frag := Fragment{"1", "user", time.Now(), time.Now(), "testType", 1, "testFragment.txt"}
		assert.Equal(t, true, WriteFragment(&frag))
	})

	t.Run("TestReadFragmentSuccess", func(t *testing.T) {
		userId := "user"
		fragmentId := "1"
		frag := Fragment{fragmentId, userId, time.Now(), time.Now(), "testType", 1, "testFragment.txt"}
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
		tempFile, _ := os.CreateTemp("", "*")
		defer os.Remove(tempFile.Name())
		tempFile.Write([]byte("Example File"))
		ok := WriteFragmentData(userId, fragmentId, tempFile)
		assert.Equal(t, true, ok)
	})

	t.Run("TestReadFragmentData", func(t *testing.T) {
		userId := "user"
		fragmentId := "1"
		tempFile, _ := os.CreateTemp("", "*")
		defer os.Remove(tempFile.Name())
		ok := WriteFragmentData(userId, fragmentId, tempFile)
		if !ok {
			t.Error("Failed to write fragment data")
			return
		}
		_, ok = ReadFragmentData(userId, fragmentId)
		assert.Equal(t, true, ok)
	})

	t.Run("TestDeleteFragmentDB", func(t *testing.T) {
		frag := CreateTestFragment()
		ok := WriteFragment(&frag)
		if !ok {
			t.Error("Failed to create fragment")
			return
		}
		tempFile, _ := os.CreateTemp("", "*")
		defer os.Remove(tempFile.Name())
		tempFile.Write([]byte("Sample data"))
		ok = WriteFragmentData(frag.OwnerId, frag.Id, tempFile)
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
}
