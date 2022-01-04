package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadHandler_uploadGetHandler(t *testing.T) {
	_, err := createTestUploadDir(t)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := &UploadHandler{
		HostAddr:  "localhost:8080",
		UploadDir: "upload_test",
	}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "testfile.txt 0 .txt\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	if err := removeTestUploadDir(t); err != nil {
		t.Errorf("removing test files: %v", err)
	}
}

func TestUploadHandler_uploadPostHandler(t *testing.T) {
	_, err := createTestUploadDir(t)
	if err != nil {
		t.Fatal(err)
	}

	fn, err := createTestFile(t)
	if err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	file, err := os.Open(fn)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()

	uploadHandler := &UploadHandler{
		UploadDir: "upload_test",
		HostAddr:  ts.URL,
	}

	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `testfile`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	if err := removeTestFiles(t); err != nil {
		t.Errorf("removing test files: %v", err)
	}

	if err := removeTestUploadDir(t); err != nil {
		t.Errorf("removing test files: %v", err)
	}
}

func removeTestFiles(t *testing.T) error {
	dir, err := os.Getwd()
	if err != nil {
		return errors.New("couldnt get dir")
	}
	test_dir := dir + "/test"

	err = os.RemoveAll(test_dir)
	if err != nil {
		return errors.New("couldnt remove test dir")
	}
	return nil
}

func createTestFile(t *testing.T) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.New("couldnt get dir")
	}

	err = os.Mkdir("test", os.ModePerm)
	if err != nil {
		return "", errors.New("couldnt make dir")
	}

	test_dir := dir + "/test"

	f1, err := os.Create(test_dir + "/testfile.txt")
	if err != nil {
		return "", errors.New("couldnt create file")
	}
	defer f1.Close()

	return f1.Name(), nil
}

func createTestUploadDir(t *testing.T) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.New("couldnt get dir")
	}

	err = os.Mkdir("upload_test", os.ModePerm)
	if err != nil {
		return "", errors.New("couldnt make dir")
	}

	test_dir := dir + "/upload_test"

	f1, err := os.Create(test_dir + "/testfile.txt")
	if err != nil {
		return "", errors.New("couldnt create file")
	}
	defer f1.Close()

	return f1.Name(), nil
}

func removeTestUploadDir(t *testing.T) error {
	dir, err := os.Getwd()
	if err != nil {
		return errors.New("couldnt get dir")
	}
	test_dir := dir + "/upload_test"

	err = os.RemoveAll(test_dir)
	if err != nil {
		return errors.New("couldnt remove test dir")
	}
	return nil
}
