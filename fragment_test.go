package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFragment(t *testing.T) {
	teardown := PreTestSetup()
	defer teardown()

	t.Run("TestGetJson", func(t *testing.T) {
		frag := CreateTestFragment()
		correctJson, _ := json.Marshal(frag)
		retrievedJson, _ := frag.GetJson()
		assert.Equal(t, string(correctJson), retrievedJson)
	})

	t.Run("TestSetData", func(t *testing.T) {
		frag := CreateTestFragment()
		ok := frag.SetData([]byte("Sample data"))
		assert.Equal(t, true, ok)
	})

	t.Run("TestGetData", func(t *testing.T) {
		frag := CreateTestFragment()
		data := []byte("Sample data")
		frag.SetData(data)
		retrievedData, ok := frag.GetData()
		assert.Equal(t, true, ok)
		assert.Equal(t, retrievedData, data)
	})

	t.Run("TestGetUserFragmentIds", func(t *testing.T) {
		frag := CreateTestFragment()
		frag.Save()
		assert.Equal(t, []string{frag.Id}, GetUserFragmentIds(frag.OwnerId))
	})

	t.Run("TestGetFragment", func(t *testing.T) {
		frag := CreateTestFragment()
		frag.Save()
		retrievedFrag, ok := GetFragment(frag.OwnerId, frag.Id)
		assert.Equal(t, true, ok)
		assert.Equal(t, frag, retrievedFrag)
	})

	t.Run("TestDeleteFragment", func(t *testing.T) {
		frag := CreateTestFragment()
		frag.Save()
		assert.Equal(t, true, DeleteFragment(frag.OwnerId, frag.Id))
	})

	t.Run("TestSave", func(t *testing.T) {
		frag := CreateTestFragment()
		assert.Equal(t, true, frag.Save())
	})

	t.Run("TestFormatsForMd", func(t *testing.T) {
		frag := CreateTestFragment()
		frag.FragmentType = "text/md"
		formats := frag.Formats()
		assert.Equal(t, 3, len(formats))
		assert.Contains(t, formats, "text/html")
		assert.Contains(t, formats, "text/md")
		assert.Contains(t, formats, "text/markdown")
	})
}
