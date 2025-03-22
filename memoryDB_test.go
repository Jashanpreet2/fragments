package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryDB(t *testing.T) {
	teardown := PreTestSetup("debug")
	defer teardown()
	t.Run("TestPutValue", func(t *testing.T) {
		testDB := localdb{}
		assert.Equal(t, true, testDB.Put("TestPK", "TestSK", "TestValue"))
	})

	t.Run("TestGetValueSuccess", func(t *testing.T) {
		testDB := localdb{}
		pk := "TestPK"
		sk := "TestSK"
		value := "TestValue"
		testDB.Put(pk, sk, value)
		returnedValue, ok := testDB.GetValue(pk, sk)
		if !ok {
			t.Fatal("Failed to retrieve value from database!")
		}
		assert.Equal(t, value, returnedValue)
	})

	t.Run("TestGetValueFail", func(t *testing.T) {
		testDB := localdb{}
		pk := "TestPK"
		sk := "TestSK"

		// Get value even though we didn't put any value
		_, ok := testDB.GetValue(pk, sk)
		assert.Equal(t, false, ok)
	})

	t.Run("TestDeleteSuccess", func(t *testing.T) {
		pk, sk, value := "TestPK", "TestSK", "TestValue"
		testDB := localdb{}
		testDB.Put(pk, sk, value)
		assert.Equal(t, true, testDB.DeleteValue(pk, sk))
	})

	t.Run("TestDeleteFail", func(t *testing.T) {
		pk, sk := "TestPK", "TestSK"
		testDB := localdb{}
		assert.Equal(t, false, testDB.DeleteValue(pk, sk))
	})

	t.Run("TestGetSKs", func(t *testing.T) {
		testDB := localdb{}
		pk, sk, value := "TestPK", "TestSK", "TestValue"
		testDB.Put(pk, sk, value)
		assert.Equal(t, []string{sk}, testDB.GetSKs(pk))
	})
}
