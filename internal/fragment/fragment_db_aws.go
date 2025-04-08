package fragment

import (
	"fmt"

	"github.com/Jashanpreet2/fragments/internal/logger"
	"github.com/Jashanpreet2/fragments/internal/memorydb"
)

var fragmentDB memorydb.LocalDB
var dataDB memorydb.LocalDB

func WriteFragment(frag *Fragment) error {
	client, err := GetDynamoDBClient()
	if err != nil {
		return err
	}
	client.WriteFragment(frag)
	return err
}

func ReadFragment(userid string, fragment_id string) (*Fragment, error) {
	client, err := GetDynamoDBClient()
	if err != nil {
		return nil, err
	}
	return client.GetFragment(userid, fragment_id)
}

func WriteFragmentData(userid string, fragment_id string, data []byte) error {
	client, err := GetS3Client()
	if err != nil {
		return err
	}
	return client.UploadFragmentDataToS3(userid, fragment_id, data)
}

func ReadFragmentData(userid string, fragment_id string) ([]byte, error) {
	client, err := GetS3Client()
	if err != nil {
		return nil, err
	}
	return client.GetFragmentDataFromS3(userid, fragment_id)
}

// Deletes the fragment metadata and data from the databases
func DeleteFragmentDB(userid string, fragment_id string) bool {
	ok := dataDB.DeleteValue(userid, fragment_id)
	if !ok {
		logger.Sugar.Error(fmt.Sprintf("Attempt to fragment data that doesn't exist with userid: %s and fragment_id: %s", userid, fragment_id))
		return false
	}
	ok = fragmentDB.DeleteValue(userid, fragment_id)
	if !ok {
		logger.Sugar.Error(fmt.Sprintf("Successfully deleted fragment data but failed to find the"+
			"fragment metadata for userid: %s and fragment_id: %s", userid, fragment_id))
		return false
	}
	return true
}

func GenerateID(userid string) int {
	// Make sure to hash the userid before passing it here else you will keep overriding the same fragment
	return len(fragmentDB.GetSKs(userid))
}

func ListFragmentIDs(userid string) ([]string, error) {
	client, err := GetDynamoDBClient()
	if err != nil {
		return nil, err
	}
	return client.GetFragmentIds(userid)
}

func ListFragmentMetadatas(userid string) []Fragment {
	fragment_ids := fragmentDB.GetSKs(userid)
	fragments := make([]Fragment, len(fragment_ids))
	i := 0
	for _, fragment_id := range fragment_ids {
		value, ok := fragmentDB.GetValue(userid, fragment_id)
		if !ok {
			logger.Sugar.Error(fmt.Sprintf("Failed to retrieve fragment with userid: %s and fragment_id: %s", userid, fragment_id))
			return fragments
		}
		fragment, ok := value.(Fragment)
		if !ok {
			logger.Sugar.Error(fmt.Sprintf("Value stored is not a Fragment for userid: %s and fragment_id: %s", userid, fragment_id))
			return fragments
		}
		fragments[i] = fragment
		i += 1
	}
	return fragments
}

func ResetDB() {
	fragmentDB = memorydb.LocalDB{}
	dataDB = memorydb.LocalDB{}
}
