package binanceUsdm

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	lvn "github.com/Lavina-Tech-LLC/lavinagopackage/v2"
	"github.com/msiddikov/cryptoposter/types"
)

type (
	reqParams struct {
		Method     string
		Endpoint   string
		ForceQuery bool
		QParams    []queryParam
	}

	queryParam struct {
		Key   string
		Value string
	}

	resp[T any] struct {
		Req     *http.Request
		Res     *http.Response
		RawBody []byte
		Resp    T
	}
)

func getHost(c *types.Auth) string {
	return lvn.Ternary(
		c.Testnet,
		"https://testnet.binancefuture.com",
		"https://fapi.binance.com")
}

func fetch[T any](c *types.Auth, r reqParams) (resp[T], error) {
	result := resp[T]{}
	endpnt := getHost(c) + r.Endpoint
	client := &http.Client{}
	if r.Method == "" {
		r.Method = "GET"
	}

	//New request
	req, err := http.NewRequest(r.Method, endpnt, nil)

	if err != nil {
		return result, err
	}

	// Params
	qparam := []queryParam{}
	bodyparam := []queryParam{}
	if r.Method == "GET" || r.ForceQuery {
		qparam = r.QParams
	} else {
		bodyparam = r.QParams
	}

	qparam = append(qparam, queryParam{
		Key:   "timestamp",
		Value: fmt.Sprint(time.Now().UnixMilli()),
	})

	QUrl := url.URL{}
	q := QUrl.Query()
	for _, v := range qparam {
		q.Add(v.Key, v.Value)
	}
	rawQuery := strings.Replace(q.Encode(), "%40", "@", -1)

	QUrl = url.URL{}
	q = QUrl.Query()
	for _, v := range bodyparam {
		q.Add(v.Key, v.Value)
	}
	bodyRawQuery := strings.Replace(q.Encode(), "%40", "@", -1)

	// Signature
	h := hmac.New(sha256.New, []byte(c.Secret))
	h.Write([]byte(rawQuery + bodyRawQuery))
	signature := hex.EncodeToString(h.Sum(nil))
	rawQuery += "&signature=" + signature

	req.URL.RawQuery = rawQuery
	req.Body = io.NopCloser(strings.NewReader(bodyRawQuery))

	// Headers
	req.Header.Add("X-MBX-APIKEY", c.Key)

	// Do request
	result.Req = req
	res, err := client.Do(req)
	if err != nil {
		return result, err
	}

	result.Res = res

	// Getting body
	body, err := io.ReadAll(res.Body)
	result.RawBody = body
	if err != nil {
		return result, err
	}

	// Status check
	if res.StatusCode == 429 {
		lvn.Logger.Noticef("BINANCE > %s %s: HTTP error: %v %s ", r.Method, r.Endpoint, res.StatusCode, string(body))
		time.Sleep(10 * time.Second)
		return fetch[T](c, r)
	}

	if res.StatusCode > 299 {
		return result, fmt.Errorf("BINANCE > %s %s: HTTP error: %v %s ", r.Method, r.Endpoint, res.StatusCode, string(body))
	}

	// Parsing body
	err = json.Unmarshal(result.RawBody, &result.Resp)
	if err != nil {
		return result, err
	}

	return result, nil
}
