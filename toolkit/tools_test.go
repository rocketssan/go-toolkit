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
