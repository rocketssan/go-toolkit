package toolkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestTool_PushJSONToRemote(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test Request Parameters
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("OK")),
			Header:     make(http.Header),
		}
	})

	var testTool Tools
	var foo struct {
		Bar string `json:"bar"`
	}
	foo.Bar = "aiueo"

	_, _, err := testTool.PushJSONToRemote("http://example.com/some/path", http.MethodPost, foo, client)
	if err != nil {
		t.Error("failed to call remote url:", err)
	}
}

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string returned")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{
		name:          "allowed no rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    false,
		errorExpected: false,
	},
	{
		name:          "allowed rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    true,
		errorExpected: false,
	},
	{
		name:          "not allowed",
		allowedTypes:  []string{"image/jpeg"},
		renameFile:    false,
		errorExpected: true,
	},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// set up a pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			// create the form data field 'file'
			part, err := writer.CreateFormFile("file", "./testdata/image.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/image.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}

		}()

		// read from the pipe which receives data
		request := httptest.NewRequest(http.MethodPost, "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTool Tools

		uploadedFiles, err := testTool.UploadFiles(request, "./testdata/uploads", e.allowedTypes, e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			filePath := fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("%s: extepted file to exist: %s", e.name, err.Error())
			}

			// clean up
			_ = os.Remove(filePath)
		}

		if e.errorExpected && err == nil {
			t.Errorf("%s: error exptected but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	// set up a pipe to avoid buffering
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()

		// create the form data field 'file'
		part, err := writer.CreateFormFile("file", "./testdata/image.png")
		if err != nil {
			t.Error(err)
		}

		f, err := os.Open("./testdata/image.png")
		if err != nil {
			t.Error(err)
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			t.Error("error decoding image", err)
		}

		err = png.Encode(part, img)
		if err != nil {
			t.Error(err)
		}

	}()

	// read from the pipe which receives data
	request := httptest.NewRequest(http.MethodPost, "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTool Tools

	uploadedFile, err := testTool.UploadOneFile(request, "./testdata/uploads", []string{"image/jpg", "image/png"})
	if err != nil {
		t.Error(err)
	}

	filePath := fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("extepted file to exist: %s", err.Error())
	}

	// clean up
	_ = os.Remove(filePath)
}

func TestTools_CreateDirIfNotExist(t *testing.T) {
	var testTool Tools

	err := testTool.CreateDirIfNotExist("./testdata/created")
	if err != nil {
		t.Error(err)
	}

	err = testTool.CreateDirIfNotExist("./testdata/created")
	if err != nil {
		t.Error(err)
	}

	_ = os.Remove("./testdata/created")
}

var slugTest = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{
		name:          "valid string",
		s:             "now is the time",
		expected:      "now-is-the-time",
		errorExpected: false,
	},
	{
		name:          "empty string",
		s:             "",
		expected:      "now-is-the-time",
		errorExpected: true,
	},
	{
		name:          "complex string",
		s:             "Now is the TIM3 to go to TH3 Marse. 123+%&",
		expected:      "now-is-the-tim3-to-go-to-th3-marse-123",
		errorExpected: false,
	},
	{
		name:          "japanese string",
		s:             "こんちわ",
		expected:      "",
		errorExpected: true,
	},
	{
		name:          "japanese and roman string",
		s:             "こんちわ hellow world",
		expected:      "hellow-world",
		errorExpected: false,
	},
}

func TestTools_Slugify(t *testing.T) {
	var testTools Tools

	for _, e := range slugTest {
		slug, err := testTools.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: error received when none exptected: %s", e.name, err.Error())
		}

		if !e.errorExpected && slug != e.expected {
			t.Errorf("%s: wrong slug returned, exptected %s but %s", e.name, e.expected, slug)
		}

		if e.errorExpected && err == nil {
			t.Errorf("%s: error exptected but none received", e.name)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	var testTools Tools

	testTools.DownloadStaticFile(rr, req, "./testdata", "image.png", "person.png")

	res := rr.Result()
	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "15980" {
		t.Error("wrong content length of", res.Header["Content-Length"][0])
	}

	if res.Header["Content-Disposition"][0] != "attachement; filename=\"person.png\"" {
		t.Error("wrong content disposition")
	}

	_, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
}

var jsonTests = []struct {
	name          string
	json          string
	errorExpected bool
	maxSize       int
	allowUnknown  bool
}{
	{
		name:          "good json",
		json:          `{"foo": "bar"}`,
		errorExpected: false,
		maxSize:       1024,
		allowUnknown:  false,
	},
	{
		name:          "badly formatted json",
		json:          `{"foo": }`,
		errorExpected: true,
		maxSize:       1024,
		allowUnknown:  false,
	},
	{
		name:          "incorrect type json",
		json:          `{"foo": 1}`,
		errorExpected: true,
		maxSize:       1024,
		allowUnknown:  false,
	},
	{
		name:          "two jsons",
		json:          `{"foo": "bar"}{"alpha": "beta"}`,
		errorExpected: true,
		maxSize:       1024,
		allowUnknown:  false,
	},
	{
		name:          "empty json",
		json:          ``,
		errorExpected: true,
		maxSize:       1024,
		allowUnknown:  false,
	},
	{
		name:          "syntax error json",
		json:          `{"foo": "bar"`,
		errorExpected: true,
		maxSize:       1024,
		allowUnknown:  false,
	},
	{
		name:          "unknown field json",
		json:          `{"food": "bar"}`,
		errorExpected: true,
		maxSize:       1024,
		allowUnknown:  false,
	},
	{
		name:          "allow unknown field json",
		json:          `{"food": "bar"}`,
		errorExpected: false,
		maxSize:       1024,
		allowUnknown:  true,
	},
	{
		name:          "missing field json",
		json:          `{aiueo: "bar"}`,
		errorExpected: true,
		maxSize:       1024,
		allowUnknown:  true,
	},
	{
		name:          "too large json",
		json:          `{"foo": "bar"}`,
		errorExpected: true,
		maxSize:       1,
		allowUnknown:  false,
	},
	{
		name:          "not json",
		json:          `Hello world`,
		errorExpected: true,
		maxSize:       5,
		allowUnknown:  false,
	},
}

func TestTool_ReadJSON(t *testing.T) {
	var testTool Tools

	for _, e := range jsonTests {
		testTool.MaxJSONSize = e.maxSize
		testTool.AllowUnknownFields = e.allowUnknown

		var decodedJSON struct {
			Foo string `json:"foo"`
		}

		// create a request with the body
		req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(e.json)))
		if err != nil {
			t.Log("Error:", err)
		}
		defer req.Body.Close()

		rr := httptest.NewRecorder()

		err = testTool.ReadJSON(rr, req, &decodedJSON)

		if e.errorExpected && err == nil {
			t.Errorf("%s: error expected, but none received", e.name)
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error not expected, but one recieved: %s", e.name, err.Error())
		}
	}
}

func TestTools_WriteJSON(t *testing.T) {
	var testTool Tools

	rr := httptest.NewRecorder()
	payload := JSONResponse{
		Error:   false,
		Message: "aiueo",
	}

	headers := make(http.Header)
	headers.Add("FOO", "BAR")

	err := testTool.WriteJSON(rr, http.StatusOK, payload, headers)
	if err != nil {
		t.Errorf("failed to write json: %v", err)
	}
}

func TestTools_ErrorJSON(t *testing.T) {
	var testTool Tools

	rr := httptest.NewRecorder()

	err := testTool.ErrorJSON(rr, errors.New("some error"), http.StatusInternalServerError)
	if err != nil {
		t.Error(err)
	}

	var payload JSONResponse
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		t.Error("received error when decoding json", err)
	}

	if !payload.Error {
		t.Error("error set to false in json, and it should be true")
	}

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("wrong status code returned. code: %d", rr.Code)
	}
}
