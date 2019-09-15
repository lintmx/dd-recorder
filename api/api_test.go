package api

import (
	"net/url"
	"testing"
)

// const testURL = "https://www.youtube.com/channel/UC1opHUrw8rvnsadT-iGp7Cg/live"

const testURL = "https://live.bilibili.com/14917277"

func TestRefreshLiveInfo(t *testing.T) {
	url, _ := url.Parse(testURL)
	api := Check(url)

	if err := api.RefreshLiveInfo(); err == nil {
		t.Logf("Success!\n\nStatus: %t\nAuthor: %s\nTitle: %s\nID: %s", api.GetLiveStatus(), api.GetAuthor(), api.GetTitle(), api.GetLiveID())
	} else {
		t.Errorf("Refresh Failed: %s", err.Error())
	}

}

func TestGetStreamURLs(t *testing.T) {
	url, _ := url.Parse(testURL)
	api := Check(url)

	if urls, err := api.GetStreamURLs(); err == nil {
		t.Log("Success")

		for _, url := range urls {
			t.Logf("Url: %s\nType: %s\n\n", url.PlayURL.String(), url.FileType)
		}
	} else {
		t.Errorf("Get Stream Urls Failed: %s", err.Error())
	}
}
