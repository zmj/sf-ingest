package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
)

const authIDcookie = "SFAPI_AuthID"

func NewUploader(host, authID string) (Uploader, error) {
	if host == "" {
		return nil, fmt.Errorf("Host empty")
	}
	if authID == "" {
		return nil, fmt.Errorf("authID empty")
	}
	cookies, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("Cookiejar init failed: %v", err)
	}
	u := &uploader{
		host:   host,
		authID: authID,
		client: &http.Client{Jar: cookies},
	}
	return u, nil
}

type uploader struct {
	host   string
	authID string
	client *http.Client
}

func (u *uploader) CreateFile(ctx context.Context, parentSfID, name string, content Content) (string, error) {
	if u == nil {
		return "", fmt.Errorf("Uploader not initialized")
	}
	for b := range content.Bytes {
		fmt.Printf("Fake upload %v\n%v", name, string(b))
	}
	return "fi123", nil
}

func (u *uploader) CreateFolder(ctx context.Context, parentSfID, name string) (string, error) {
	if u == nil {
		return "", fmt.Errorf("Uploader not initialized")
	}
	url := fmt.Sprintf("https://%v/sf/v3/Items(%v)/Folder", u.host, parentSfID)
	body := bytes.NewReader([]byte(fmt.Sprintf("{\"Name\": \"%v\"}", name)))
	req, err := http.NewRequest("POST", url, body)
	req.AddCookie(&http.Cookie{Name: authIDcookie, Value: u.authID})
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return "", fmt.Errorf("Failed to get createFolder request: %v", err)
	}

	s, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("%v\n", string(s))

	resp, err := u.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("createFolder call failed: %v", err)
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read createFolder response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("createFolder api error: %v %v", resp.Status, string(bytes))
	}
	var folder folder
	err = json.Unmarshal(bytes, &folder)
	if err != nil {
		return "", fmt.Errorf("Failed to parse createFolder response: %v %v", err, string(bytes))
	}
	if folder.ID == "" {
		return "", fmt.Errorf("CreateFolder response did not contain id: %v", string(bytes))
	}
	return folder.ID, nil
}
