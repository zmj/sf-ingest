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
	"time"
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
	url := fmt.Sprintf("https://%v/sf/v3/Items(%v)/Upload2", u.host, parentSfID)
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

	req, err := http.NewRequest("POST", us.ChunkURI+"&fmt=json", &content)
	if err != nil {
		return "", fmt.Errorf("Create upload httpReq failed: %v", err)
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	var ur uploadResult
	err = u.do(req, &ur)
	if ur.Error {
		return "", fmt.Errorf("Upload server error: %v", ur.ErrorMessage)
	}
	if len(ur.Value) == 0 {
		return "", fmt.Errorf("Upload server returned no file")
	}
	if ur.Value[0].ID == "" {
		return "", fmt.Errorf("Upload response does not contain id")
	}
	return ur.Value[0].ID, nil
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
	return u.do(req, expectedResp)
}

func (u *uploader) do(req *http.Request, expectedResp interface{}) error {
	//s, _ := httputil.DumpRequestOut(req, true)
	//fmt.Printf("%v\n", string(s))
	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("Request failed: %v", err)
	}
	//s, _ = httputil.DumpResponse(resp, true)
	//fmt.Printf("%v\n", string(s))
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

func (c *Content) Read(dst []byte) (int, error) {
	read := 0
	for len(dst) > 0 {
		fmt.Printf("src before %v\n", len(c.current))
		fmt.Printf("dst before %v\n", len(dst))
		if len(c.current) == 0 {
			next, ok := <-c.Bytes
			if !ok {
				fmt.Println("r1")
				return read, io.EOF
			}
			c.current = next
		}
		r := copy(dst, c.current)
		c.current = c.current[r:]
		dst = dst[r:]
		read += r
		fmt.Printf("src after %v\n", len(c.current))
		fmt.Printf("dst after %v\n", len(dst))
		<-time.After(time.Second)
	}
	fmt.Println("r2")
	return read, nil
}
