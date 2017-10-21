package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	defer func() {
		for range content.Bytes {
		}
	}()
	if u == nil {
		return "", fmt.Errorf("Uploader not initialized")
	}
	url := fmt.Sprintf("https://%v/sf/v3/Items(%v)/Upload2")
	usr := uploadSpecReq{
		Method:   "Standard",
		Raw:      true,
		FileName: name,
		FileSize: content.Size,
	}
	var us uploadSpec
	err := u.doApiPost(url, &usr, &us)
	if err != nil {
		return "", fmt.Errorf("Upload API call failed: %v", err)
	}

	req, err := http.NewRequest("POST", us.ChunkURI, content.reader())
	if err != nil {
		return "", fmt.Errorf("Create upload httpReq failed: %v", err)
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	resp, err := u.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Upload request failed: %v", err)
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read upload response: %v", err)
	}
	var ur uploadResult
	err = json.Unmarshal(respBytes, &ur)
	if err != nil {
		return "", fmt.Errorf("Failed to parse response: %v", err)
	}
	if ur.Error {
		return "", fmt.Errorf("Upload server error: %v", ur.ErrMsg)
	}
	if ur.Value.ID == "" {
		return "", fmt.Errorf("Upload response does not contain id")
	}
	return ur.Value.ID, nil
}

func (u *uploader) CreateFolder(ctx context.Context, parentSfID, name string) (string, error) {
	if u == nil {
		return "", fmt.Errorf("Uploader not initialized")
	}
	url := fmt.Sprintf("https://%v/sf/v3/Items(%v)/Folder", u.host, parentSfID)
	toCreate := folder{Name: name}
	var folder folder
	err := u.doApiPost(url, &toCreate, &folder)
	if err != nil {
		return "", fmt.Errorf("Create folder API call failed: %v", err)
	}
	if folder.ID == "" {
		return "", fmt.Errorf("CreateFolder response did not contain id")
	}
	return folder.ID, nil
}

func (u *uploader) doApiPost(url string, body, expectedResp interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("Failed to serialize body: %v", err)
	}
	bodyRdr := bytes.NewReader(bodyBytes)
	req, err := http.NewRequest("POST", url, bodyRdr)
	if err != nil {
		return fmt.Errorf("Failed to create httpReq: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: authIDcookie, Value: u.authID})
	req.Header.Add("Content-Type", "application/json")

	s, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("%v\n", string(s))

	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Api error: %v %v", resp.Status, string(respBytes))
	}
	err = json.Unmarshal(respBytes, expectedResp)
	if err != nil {
		return fmt.Errorf("Failed to parse response: %v %v", err, string(respBytes))
	}
	return nil
}

func (c *Content) reader() io.Reader {
	return nil
}
