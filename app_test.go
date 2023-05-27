package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
)

const (
	expTime       = 60
	longURL       = "https://www.baidu.com"
	shortLink     = "IFHZzaO"
	ShortlinkInfo = `{"url": "https://www.baidu.com", "created_at": "2019-12-12 12:12:12", "expiration": 60}`
)

type storageMock struct {
	mock.Mock
}

var app App
var mockR *storageMock

func (s *storageMock) Shorten(url string, exp int64) (string, error) {
	args := s.Called(url, exp)
	fmt.Println(args)
	return args.String(0), args.Error(1)
}

func (s *storageMock) Unshorten(eid string) (string, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func (s *storageMock) ShortlinkInfo(eid string) (interface{}, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func init() {
	app = App{}
	mockR = new(storageMock)
	app.Initialize(&Env{S: mockR})
}

func TestCreateShortlink(t *testing.T) {
	mockR.On("Shorten", longURL, int64(expTime)).Return(shortLink, nil).Once()

	var jsonStr = []byte(`{"url":"https://www.baidu.com","expiration_in_minutes":60}`)
	req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw := httptest.NewRecorder()
	app.Router.ServeHTTP(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatalf("Excepted status created, got %d", rw.Code)
	}
	resp := struct {
		Shortlink string `json:"shortlink"`
	}{}

	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("should decode the response")
	}
	if resp.Shortlink != shortLink {
		t.Fatalf("Excepted receive %s, got %s", shortLink, resp.Shortlink)
	}
}

func TestRedirect(t *testing.T) {
	r := fmt.Sprintf("/%s", shortLink)
	req, err := http.NewRequest("GET", r, nil)
	if err != nil {
		t.Fatal("Should be able to create a requrest.", err)
	}
	mockR.On("Unshorten", shortLink).Return(longURL, nil).Once()

	rw := httptest.NewRecorder()
	app.Router.ServeHTTP(rw, req)

	if rw.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Excepted redirect %d. got %d", http.StatusTemporaryRedirect, rw.Code)
	}
}
