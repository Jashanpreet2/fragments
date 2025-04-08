package fragment_test

// import (
// 	"encoding/json"
// 	"testing"

// 	"github.com/Jashanpreet2/fragments/internal/fragment"
// 	"github.com/Jashanpreet2/fragments/internal/testutils"
// 	"github.com/stretchr/testify/assert"
// )

// func TestFragment(t *testing.T) {
// 	t.Run("TestGetJson", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		correctJson, _ := json.Marshal(frag)
// 		retrievedJson, _ := frag.GetJson()
// 		assert.Equal(t, string(correctJson), retrievedJson)
// 	})

// 	t.Run("TestSetData", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		ok := frag.SetData([]byte("Sample data"))
// 		assert.Equal(t, true, ok)
// 	})

// 	t.Run("TestGetData", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		data := []byte("Sample data")
// 		frag.SetData(data)
// 		retrievedData, ok := frag.GetData()
// 		assert.Equal(t, true, ok)
// 		assert.Equal(t, retrievedData, data)
// 	})

// 	t.Run("TestGetUserFragmentIds", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		frag.Save()
// 		assert.Equal(t, []string{frag.Id}, fragment.GetUserFragmentIds(frag.OwnerId))
// 	})

// 	t.Run("TestGetFragment", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		frag.Save()
// 		retrievedFrag, ok := fragment.GetFragment(frag.OwnerId, frag.Id)
// 		assert.Equal(t, true, ok)
// 		assert.Equal(t, frag, retrievedFrag)
// 	})

// 	t.Run("TestDeleteFragment", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		frag.Save()
// 		assert.Equal(t, true, fragment.DeleteFragment(frag.OwnerId, frag.Id))
// 	})

// 	t.Run("TestSave", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		assert.Equal(t, true, frag.Save())
// 	})

// 	t.Run("TestFormatsForMd", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		frag.FragmentType = "text/md"
// 		formats := frag.Formats()
// 		assert.Equal(t, 3, len(formats))
// 		assert.Contains(t, formats, "text/html")
// 		assert.Contains(t, formats, "text/md")
// 		assert.Contains(t, formats, "text/markdown")
// 	})

// 	t.Run("TestGetDataWithoutSavingData", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		data, ok := frag.GetData()
// 		assert.Equal(t, []byte(nil), data)
// 		assert.Equal(t, false, ok)
// 	})
// }
