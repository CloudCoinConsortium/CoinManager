package api

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/kr/jsonfeed"
)

type NewsResponse struct {
  Version string `json:"version"`
  Title string `json:"title"`
  Items []NewsItem `json:"items"`
}

type NewsItem struct {
  Id string `json:"id"`
  Title string `json:"title"`
  ContentText string `json:"content_text"`
  ContentHtml string `json:"content_html"`
  Summary string `json:"summary"`

  Image string `json:"image"`
  DatePublished time.Time `json:"date_published"`
  Author *Author `json:"author"`
}

type Author struct {
  Name string `json:"name"`
  Avatar string `json:"avatar"`
  URL string
}

func NewsReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  logger.L(ctx).Debugf("News Request")
/*
  n := `{
    "version": "https://jsonfeed.org/version/1",
    "title": "CloudCoin News Feed",
    "home_page_url": "https://cloudcoinconsortium.com",
    "items": [
        {
            "id": "01172022A",
            "title": "CloudCoin Project Development is Complete!",
            "date_published": "2022-01-17T19:30:00-01:00",
            "content_html": "CloudCoin is now completeley developed"
        },
        {
            "id": "11172022B",
            "title": "SkyVault.cc is up and running",
            "date_published": "2022-01-17T19:30:00-01:00",
            "content_html": "Check it out at <a href=\"https://skywallet.cc\">Skywallet.cc</a>"
        }
    ]
}
`
*/

  resp, err := http.Get(config.NEWSFEED_URL)
  if err != nil {
    logger.L(ctx).Errorf("Failed to make request %s:%s", config.NEWSFEED_URL, err.Error())
    ErrorResponse(ctx, w, perror.New(perror.ERROR_HTTP_NEWSFEED, "Failed to make request to NewsFeed Server: " + err.Error()))
    return
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read JSON Feed %s", err.Error())
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARSE_FEED, "Failed to read JSON Feed from the NewsFeed Server " + err.Error()))
    return
  }

  jf := &jsonfeed.Feed{}
  err = jf.UnmarshalJSON(body)
  if err != nil {
    logger.L(ctx).Errorf("Failed to parse JSON Feed %s", err.Error())
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARSE_FEED, "Failed to parse JSON Feed"))
    return
  }


  SuccessResponse(ctx, w, jf)
}

