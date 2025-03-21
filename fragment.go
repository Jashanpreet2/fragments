package main

import (
	"encoding/json"
	"errors"
	"mime"
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
	Size         int
}

func (frag *Fragment) GetJson() (string, bool) {
	jsonData, err := json.Marshal(frag)
	if err != nil {
		return "", false
	}
	return string(jsonData), true
}

func (frag *Fragment) GetData() ([]byte, bool) {
	file, ok := ReadFragmentData(frag.OwnerId, frag.Id)
	if !ok {
		sugar.Errorf("Failed to find data for the current fragment at userid: %s and fragment_id: %s", frag.OwnerId, frag.Id)
		return nil, false
	}
	return file, true
}

func (frag *Fragment) SetData(data []byte) bool {
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

func (frag *Fragment) ConvertMimetype(ext string) ([]byte, string, error) {
	data, ok := frag.GetData()
	mime.AddExtensionType(".md", "text/markdown")
	mime.AddExtensionType(".markdown", "text/markdown")
	mimeType := strings.Split(mime.TypeByExtension(ext), ";")[0]
	if mimeType == "" {
		return nil, "", errors.New("extension doesn't exist")
	}
	if !ok {
		sugar.Error("Failed to get fragment data")
		return nil, "", errors.New("unable to retrieve data")
	}
	if frag.MimeType() == "text/markdown" {
		sugar.Info(mimeType)
		if mimeType == "text/html" {
			return ConvertMdToHtml(data), "text/markdown", nil
		}
	}
	return nil, "", errors.New("unsupported extension")
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
	fragment, ok := ReadFragment(username, fragment_id)
	if !ok {
		sugar.Info("Unable to find the specified fragment. Previously this error was encountered when the username wasn't hashed")
	}
	return fragment, ok
}

func DeleteFragment(username string, fragment_id string) bool {
	return DeleteFragmentDB(username, fragment_id)
}

func IsSupportedType(typename string) bool {
	supportedTypes := []string{"text/plain"}
	return slices.Contains(supportedTypes, strings.Split(typename, ";")[0])
}
