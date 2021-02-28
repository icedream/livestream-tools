package metacollector

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type MetaCollectorClient struct {
	client *http.Client
	apiURL *url.URL
}

type MetaCollectorResponse struct {
	Artist, Title, Publisher string
	CoverURL                 *string
}

type MetaCollectorRequest struct {
	Artist, Title string
}

func NewMetaCollectorClient(apiURL *url.URL) *MetaCollectorClient {
	return &MetaCollectorClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		apiURL: apiURL,
	}
}

func (mcc *MetaCollectorClient) json(path string, data interface{}, responseData interface{}) (err error) {
	u := mcc.apiURL.ResolveReference(&url.URL{
		Path: path,
	})
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(data); err != nil {
		return
	}
	resp, err := mcc.client.Post(u.String(), "application/json", buf)
	if err != nil {
		return
	}
	err = json.NewDecoder(resp.Body).Decode(responseData)
	return
}

func (mcc *MetaCollectorClient) path(parts ...string) string {
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func (mcc *MetaCollectorClient) GetTrack(req MetaCollectorRequest) (resp *MetaCollectorResponse, err error) {
	resp = new(MetaCollectorResponse)
	err = mcc.json("track/find", req, resp)
	return
}
