package library

import (
	"context"
	"github.com/Ian-zy0329/go-mall/common/util/httptool"
	"github.com/Ian-zy0329/go-mall/library"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)
	httptool.SetUTHttpClient(client)
	os.Exit(m.Run())
}

func TestWhoisLib_GetHostDetail(t *testing.T) {
	defer gock.Off()
	gock.New("https://ipwho.is").
		MatchHeader("User-Agent", "curl/7.77.0").Get("").
		Reply(200).
		BodyString("{\"ip\":\"127.126.113.220\",\"success\":true}")
	ipDetail, err := library.NewWhoisLib(context.TODO()).GetHostIpDetail()
	assert.Nil(t, err)
	assert.Equal(t, ipDetail.Ip, "127.126.113.220")
}
