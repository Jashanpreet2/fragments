package main

import (
	"fmt"
)

var fragmentDB localdb
var dataDB localdb

func WriteFragment(frag *Fragment) bool {
	return fragmentDB.Put(frag.OwnerId, frag.Id, *frag)
}

func ReadFragment(userid string, fragment_id string) (Fragment, bool) {
	value, ok := fragmentDB.GetValue(userid, fragment_id)
	if ok {
		frag, ok := value.(Fragment)
		if ok {
			return frag, true
		} else {
			sugar.Infof("No fragment found! User ID: %s. Fragment ID: %s.", userid, fragment_id)
		}
	}
	return Fragment{}, false
}

func WriteFragmentData(userid string, fragment_id string, data []byte) bool {
	dataDB.Put(userid, fragment_id, data)
	return true
}

func ReadFragmentData(userid string, fragment_id string) ([]byte, bool) {
	fragmentData, ok := dataDB.GetValue(userid, fragment_id)
	if !ok {
		sugar.Infof("Failed to find fragment data for User ID: %s and Fragment ID: %s", userid, fragment_id)
		return nil, false
	}
	fragment_file, ok := fragmentData.([]byte)
	if !ok {
		sugar.Infof("Fragment data not of type []byte for User ID: %s and Fragment ID: %s", userid, fragment_id)
		return nil, false
	}
	return fragment_file, true
}

// Deletes the fragment metadata and data from the databases
func DeleteFragmentDB(userid string, fragment_id string) bool {
	ok := dataDB.DeleteValue(userid, fragment_id)
	if !ok {
		sugar.Error(fmt.Sprintf("Attempt to fragment data that doesn't exist with userid: %s and fragment_id: %s", userid, fragment_id))
		return false
	}
	ok = fragmentDB.DeleteValue(userid, fragment_id)
	if !ok {
		sugar.Error(fmt.Sprintf("Successfully deleted fragment data but failed to find the"+
			"fragment metadata for userid: %s and fragment_id: %s", userid, fragment_id))
		return false
	}
	return true
}

func GenerateID(userid string) int {
	return len(fragmentDB.GetSKs(userid))
}

func ListFragmentIDs(userid string) []string {
	return fragmentDB.GetSKs(userid)
}

func ListFragmentMetadatas(userid string) []Fragment {
	fragment_ids := fragmentDB.GetSKs(userid)
	fragments := make([]Fragment, len(fragment_ids))
	i := 0
	for _, fragment_id := range fragment_ids {
		value, ok := fragmentDB.GetValue(userid, fragment_id)
		if !ok {
			sugar.Error(fmt.Sprintf("Failed to retrieve fragment with userid: %s and fragment_id: %s", userid, fragment_id))
			return fragments
		}
		fragment, ok := value.(Fragment)
		if !ok {
			sugar.Error(fmt.Sprintf("Value stored is not a Fragment for userid: %s and fragment_id: %s", userid, fragment_id))
			return fragments
		}
		fragments[i] = fragment
		i += 1
	}
	return fragments
}
