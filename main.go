package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"github.com/hailiang/gosocks"
)

type Config struct {
	SocksProxyAddress string
	Username string
	Password string
	APIHost string
	SourceApp string
}

type AuthResponse struct {
	Token string `json:"token"`
	UserId int `json:"userId"`
}

type QuoteRequest struct {
	XMLName xml.Name `xml:"getQuotes"`
	Items []QuoteItem `xml:"item"`
}
type QuoteResponse struct {
	XMLName xml.Name `xml:""ns2:getQuotesResponse""`
	Items []QuoteItem `xml:"item"`
}

type OptionChainRequest struct {
	XMLName xml.Name `xml:"getOptionChain"`
	Symbols []string `xml:"symbol"`
}

type OptionChainResponse struct {
	XMLName xml.Name `xml:""ns2:getQuotesResponse""`
	Items []QuoteOptionItem `xml:"item"`
}

type QuotePrice struct {
	Amount float64 `xml:"amount"`
	Currency string `xml:"currency"`
}

type QuoteItem struct {
	XMLName xml.Name `xml:"item"`
	AskPrice QuotePrice `xml:"askPrice"`
	AskSize float64 `xml:"askSize"`
	BidPrice QuotePrice`xml:"bidPrice"`
	BidSize float64 `xml:"bidSize"`
	ClosingMark QuotePrice`xml:"closingMark"`
	HighPrice QuotePrice`xml:"highPrice"`
	ImpliedVolatility float64 `xml:"impliedVolatility"`
	LastSize int `xml:"lastSize"`
	LastTradeTimeMillis int `xml:"lastTradeTimeMillis"`
	LastTradedPrice QuotePrice`xml:"lastTradedPrice"`
	LowPrice QuotePrice`xml:"lowPrice"`
	Mark QuotePrice`xml:"mark"`
	MarkChangePct float64 `xml:"markChangePct"`
	MarkChangePrice QuotePrice`xml:"markChangePrice"`
	OpenPrice QuotePrice`xml:"openPrice"`
	Symbol string `xml:"symbol"`
	Volume int `xml:"volume"`
	YearHighPrice QuotePrice`xml:"yearHighPrice"`
	YearLowPrice QuotePrice`xml:"yearLowPrice"`
	DivType string `xml:"divType"`
	Dividend QuotePrice `xml:"dividend"`
	DividendDate int `xml:"dividendDate"`
	ExtLastTradedPrice QuotePrice `xml:"extLastTradedPrice"`
	InstrumentType string `xml:"instrumentType"`
	PreviousClosePrice QuotePrice `xml:"previousClosePrice"`
	SaleTradeTimeMillis int `xml:"saleTradeTimeMillis"`
	TradeCondition int `xml:"tradeCondition"`
}

type QuoteOptionItem struct {
	XMLName xml.Name `xml:"item"`
	Order int `xml:"order"`
	AddFlag string `xml:"addFlag"`
	DaysToExpire int `xml:"daysToExpire"`
	ExpiryLabel string `xml:"expiryLabel"`
	ExpiryType string `xml:"expiryType"`
	OptionCollection []QuoteOptionStrikePair `xml:"option_Collection"`

}

type QuoteOptionStrikePair struct {
	XMLName xml.Name `xml:"option_Collection"`
	Call QuoteOption `xml:"call"`
	Put QuoteOption `xml:"put"`
	Strike float64 `xml:"strike"`
}
type QuoteOption struct {
	DeliverableType string `xml:"deliverableType"`
	Exchange string `xml:"exchange"`
	ExchangeType string `xml:"exchangeType"`
	ExerciseStyle string `xml:"exerciseStyle"`
	ExpirationType string `xml:"expirationType"`
	ExpireType string `xml:"expireType"`
	ExpiryDeliverable string `xml:"expiryDeliverable"`
	Instrument QuoteInstrument `xml:"instrument"`
	InstrumentId int `xml:"instrumentId"`
	MinimumTickValue1 float64 `xml:"minimumTickValue1"`
	MinimumTickValue2 float64 `xml:"minimumTickValue2"`
	Multiplier int `xml:"multiplier"`
	OpraRoot string `xml:"opraRoot"`
	ReutersInstrumentCode string `xml:"reutersInstrumentCode"`
	SharesPerContract int `xml:"sharesPerContract"`
	StrikePrice float64 `xml:"strikePrice"`
	Symbol string `xml:"symbol"`
}

type QuoteInstrument struct {
	DaysToExpire int `xml:"daysToExpire"`
	EasyToBorrow bool `xml:"easyToBorrow"`
	ExchangeCode string `xml:"exchangeCode"`
	ExchangeType string `xml:"exchangeType"`
	ExerciseStyle string `xml:"exerciseStyle"`
	ExpireDay int `xml:"expireDay"`
	ExpireDayET int `xml:"expireDayET"`
	InstrumentId int `xml:"instrumentId"`
	InstrumentSubType string `xml:"instrumentSubType"`
	InstrumentType string `xml:"instrumentType"`
	MinimumTickValue1 float64 `xml:"minimumTickValue1"`
	MinimumTickValue2 float64 `xml:"minimumTickValue2"`
	Month int `xml:"month"`
	Multiplier float64 `xml:"multiplier"`
	OpraCode string `xml:"opraCode"`
	OptionType string `xml:"optionType"`
	Optionable bool `xml:"optionable"`
	ReutersInstrumentCode string `xml:"reutersInstrumentCode"`
	StrikePrice float64 `xml:"strikePrice"`
	Symbol string `xml:"symbol"`
	Tradeable bool `xml:"tradeable"`
	UnderlyingInstrumentId int `xml:"underlyingInstrumentId"`
	UnderlyingSymbol string `xml:"underlyingSymbol"`
	Year int `xml:"year"`
}
type MonsterClient struct {
	client *http.Client
	config *Config
	account *AuthResponse
}

func (m MonsterClient) Auth() bool {
	data := url.Values{"j_username":{m.config.Username}, "j_password":{m.config.Password}}
	req_url := strings.Join([]string{"https://", m.config.APIHost, "/j_acegi_security_check"}, "")
	req, err := http.NewRequest("POST", req_url, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "text/xml,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("sourceapp", m.config.SourceApp)

	resp, err := m.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// TODO: Actually check if there was an error, and return it for handling
	fmt.Printf("Auth body was : %s\n", body)
	// TODO: Check response body for error types and return for handling
	json.Unmarshal(body, &m.account)
	fmt.Printf("Setting account token to: %s\n", m.account.Token)
	return true
}

func (m MonsterClient) Post(path string, payload []byte) (response_body []byte, err error) {
	req_url := strings.Join([]string{"https://", m.config.APIHost, path}, "")
	req, err := http.NewRequest("POST", req_url, bytes.NewReader(payload))
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("sourceapp", m.config.SourceApp)
	req.Header.Add("token", m.account.Token)
	quoteresp, err := m.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response status was: %v\n", quoteresp.StatusCode)
	defer quoteresp.Body.Close()
	body, err := ioutil.ReadAll(quoteresp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body, nil
}

func main() {

	config := Config{}
	config.SocksProxyAddress = os.Getenv("SOCKS_PROXY_ADDR")
	config.Username = os.Getenv("MONSTER_USER")
	config.Password = os.Getenv("MONSTER_PASS")
	config.APIHost = os.Getenv("MONSTER_HOST")
	config.SourceApp = os.Getenv("MONSTER_SOURCEAPP")


	dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, config.SocksProxyAddress)
	transport := &http.Transport{
		Dial: dialSocksProxy,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	options := cookiejar.Options{
			    }
	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}
	client := MonsterClient{
		client: &http.Client{Transport: transport, Jar: jar},
		config: &config,
		account: &AuthResponse{},
	}
	client.Auth()

	//var reqData QuoteRequest
	//qitem := QuoteItem{Symbol: "IBM", InstrumentType: "Equity"}
	//reqData.Items = append(reqData.Items, qitem)
	//payload, err := xml.Marshal(reqData)
	//quoteresp, _ := client.Post("/services/quotesService", payload)
	//var quotes QuoteResponse
	//xml.Unmarshal(quoteresp, &quotes)

	var optData OptionChainRequest
	optData.Symbols = append(optData.Symbols, "IBM")
	optData.Symbols = append(optData.Symbols, "AAPL")
	opayload, err := xml.Marshal(optData)
	fmt.Printf("Option request payload is: %s", opayload)
	optionsresp, _ := client.Post("/services/quotesOptionService", opayload)
	var chain OptionChainResponse
	xml.Unmarshal(optionsresp, &chain)

	//fmt.Printf("Body was : %s\n", optionsresp)
	col := chain.Items[0].OptionCollection[0]
	xbyt, _ := xml.Marshal(col)
	fmt.Printf("options chain : %v\n", col)
	fmt.Printf("xml : %s\n", string(xbyt))

}