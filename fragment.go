package main

import (
	"encoding/json"
	"mime/multipart"
	"slices"
	"strings"
	"time"
)

type Fragment struct {
	Id           string
	OwnerId      string
	Created      time.Time
	Updated      time.Time
	FragmentType string
	Size         int64
	FragmentName string
}

func (frag *Fragment) GetJson() (string, bool) {
	jsonData, err := json.Marshal(frag)
	if err != nil {
		return "", false
	}
	return string(jsonData), true
}

func (frag *Fragment) GetData() (multipart.File, bool) {
	file, ok := ReadFragmentData(frag.OwnerId, frag.Id)
	if !ok {
		sugar.Errorf("Failed to find data for the current fragment at userid: %s and fragment_id: %s", frag.OwnerId, frag.Id)
		return nil, false
	}
	return file, true
}

func (frag *Fragment) SetData(data multipart.File) bool {
	frag.Updated = time.Now()
	WriteFragment(frag)
	return WriteFragmentData(frag.OwnerId, frag.Id, data) && WriteFragment(frag)
}

func (frag *Fragment) Save() bool {
	return WriteFragment(frag)
}

func (frag *Fragment) MimeType() string {
	return strings.Split(frag.FragmentType, ";")[0]
}

func (frag *Fragment) Formats() []string {
	mimeType := frag.MimeType()
	if mimeType == "text/plain" {
		return []string{"text/plain"}
	}
	return []string{mimeType}
}

func GetUserFragmentIds(username string) []string {
	return ListFragmentIDs(username)
}

func GetFragment(username string, fragment_id string) (Fragment, bool) {
	return ReadFragment(username, fragment_id)
}

func DeleteFragment(username string, fragment_id string) bool {
	return DeleteFragmentDB(username, fragment_id)
}

func IsSupportedType(typename string) bool {
	supportedTypes := []string{"text/plain"}
	return slices.Contains(supportedTypes, strings.Split(typename, ";")[0])
}
