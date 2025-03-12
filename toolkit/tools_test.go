package toolkit

import (
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
