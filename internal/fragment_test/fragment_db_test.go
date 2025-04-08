package fragment_test

// import (
// 	"testing"
// 	"time"

// 	"github.com/Jashanpreet2/fragments/internal/fragment"
// 	"github.com/Jashanpreet2/fragments/internal/testutils"
// 	"github.com/stretchr/testify/assert"
// )

// func TestFragmentDBInterface(t *testing.T) {
// 	t.Run("TestWriteFragment", func(t *testing.T) {
// 		frag := fragment.Fragment{Id: "1", OwnerId: "user", Created: time.Now(), Updated: time.Now(), FragmentType: "testType", Size: 1}
// 		assert.Equal(t, true, fragment.WriteFragment(&frag))
// 	})

// 	t.Run("TestReadFragmentSuccess", func(t *testing.T) {
// 		userId := "user"
// 		fragmentId := "1"
// 		frag := fragment.Fragment{Id: fragmentId, OwnerId: userId, Created: time.Now(), Updated: time.Now(), FragmentType: "testType", Size: 1}
// 		ok := assert.Equal(t, true, fragment.WriteFragment(&frag))
// 		if !ok {
// 			t.Error("Failed to create fragment")
// 			return
// 		}
// 		_, ok = fragment.ReadFragment(userId, fragmentId)
// 		assert.Equal(t, true, ok)
// 	})

// 	t.Run("TestReadFragmentFail", func(t *testing.T) {
// 		_, ok := fragment.ReadFragment("Non existent User ID", "Non existentfragment.Fragment ID")
// 		assert.Equal(t, false, ok)
// 	})

// 	t.Run("TestWriteFragmentData", func(t *testing.T) {
// 		userId := "user"
// 		fragmentId := "1"
// 		ok := fragment.WriteFragmentData(userId, fragmentId, []byte("Sample data"))
// 		assert.Equal(t, true, ok)
// 	})

// 	t.Run("TestReadFragmentData", func(t *testing.T) {
// 		userId := "user"
// 		fragmentId := "1"
// 		data := []byte("Sample data")
// 		if ok := fragment.WriteFragmentData(userId, fragmentId, data); !ok {
// 			t.Error("Failed to write user fragment")
// 		}
// 		retrievedData, ok := fragment.ReadFragmentData(userId, fragmentId)
// 		assert.Equal(t, ok, true)
// 		assert.Equal(t, data, retrievedData)
// 	})

// 	t.Run("TestDeleteFragmentDb", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		data := []byte("Sample data")
// 		ok := fragment.WriteFragmentData(frag.OwnerId, frag.Id, data)
// 		if !ok {
// 			t.Error("Failed to write fragment data")
// 			return
// 		}
// 		ok = fragment.DeleteFragmentDB(frag.OwnerId, frag.Id)
// 		assert.Equal(t, true, ok)
// 	})

// 	t.Run("TestListFragmentIds", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		ok := fragment.WriteFragment(&frag)
// 		if !ok {
// 			t.Error("Failed to write fragment")
// 		}
// 		ids := fragment.GetUserFragmentIds(frag.OwnerId)
// 		assert.Equal(t, []string{frag.Id}, ids)
// 	})

// 	t.Run("TestListFragmentMetadatas", func(t *testing.T) {
// 		frag := testutils.CreateTestFragment()
// 		ok := fragment.WriteFragment(&frag)
// 		if !ok {
// 			t.Error("Failed to write fragment")
// 		}
// 		retrievedFrags := fragment.ListFragmentMetadatas(frag.OwnerId)
// 		assert.Equal(t, []fragment.Fragment{frag}, retrievedFrags)
// 	})

// 	t.Run("TestGenerateID", func(t *testing.T) {
// 		assert.Equal(t, 0, fragment.GenerateID("Non-existent user"))
// 	})

// 	t.Run("Testfragment.ReadFragmentDataFails", func(t *testing.T) {
// 		_, ok := fragment.ReadFragmentData("NonExistentUser", "NonExistentId")
// 		assert.Equal(t, false, ok)
// 	})

// 	t.Run("TestDeleteFragmentDbDataDoesn'tExist", func(t *testing.T) {
// 		fragment.ResetDB()
// 		frag := testutils.CreateTestFragment()
// 		fragment.WriteFragment(&frag)
// 		ok := fragment.DeleteFragmentDB(frag.OwnerId, frag.Id)
// 		assert.Equal(t, false, ok)
// 	})

// 	t.Run("TestDeleteFragmentDbFragmentDoesn'tExist", func(t *testing.T) {
// 		fragment.ResetDB()
// 		frag := testutils.CreateTestFragment()
// 		fragment.WriteFragmentData(frag.OwnerId, frag.Id, []byte("Some data"))
// 		ok := fragment.DeleteFragmentDB(frag.OwnerId, frag.Id)
// 		assert.Equal(t, false, ok)
// 	})
// }
