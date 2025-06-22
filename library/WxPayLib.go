package library

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/common/util/httptool"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/Ian-zy0329/go-mall/resources"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type WxPayLib struct {
	ctx       context.Context
	payConfig WxPayConfig
}

type WxPayConfig struct {
	AppId           string
	MchId           string
	PrivateSerialNo string
	AesKey          string
	NotifyUrl       string
}

func NewWxPayLib(ctx context.Context, payConfig WxPayConfig) *WxPayLib {
	return &WxPayLib{
		ctx:       ctx,
		payConfig: payConfig,
	}
}

const prePayApiUrl = "https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi"

type PrePayParam struct {
	AppId       string `json:"app_id"`
	MchId       string `json:"mch_id"`
	Description string `json:"description"`
	OutTradeNo  string `json:"out_trade_no"`
	NotifyUrl   string `json:"notify_url"`
	Amount      struct {
		Total    int    `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Payer struct {
		OpenId string `json:"open_id"`
	} `json:"payer"`
}

type WxPayInvokeInfo struct {
	AppId     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

type WxPayNotifyResponse struct {
	CreateTime string              `json:"create_time"`
	Resource   WxPayNotifyResource `json:"resource"`
}

type WxPayNotifyResource struct {
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
}

type WxPayNotifyResourceData struct {
	TransactionID string `json:"transaction_id"`
	Amount        struct {
		PayerTotal    int    `json:"payer_total"`
		Total         int    `json:"total"`
		Currency      string `json:"currency"`
		PayerCurrency string `json:"payer_currency"`
	} `json:"amount"`
	Mchid       string    `json:"mchid"`
	TradeState  string    `json:"trade_state"`
	BankType    string    `json:"bank_type"`
	SuccessTime time.Time `json:"success_time"`
	Payer       struct {
		Openid string `json:"openid"`
	} `json:"payer"`
	OutTradeNo     string `json:"out_trade_no"`
	AppId          string `json:"AppID"`
	TradeStateDesc string `json:"trade_state_desc"`
	TradeType      string `json:"trade_type"`
	Attach         string `json:"attach"`
}

func (wpl *WxPayLib) CreateOrderPay(order *do.Order, userOpenId string) (payInvokeInfo *WxPayInvokeInfo, err error) {
	payDescription := fmt.Sprintf("GOMALL 商场购买%s等商品", order.Items[0].CommodityName)
	prePayParam := &PrePayParam{
		AppId:       wpl.payConfig.AppId,
		MchId:       wpl.payConfig.MchId,
		Description: payDescription,
		OutTradeNo:  order.OrderNo,
		NotifyUrl:   wpl.payConfig.NotifyUrl,
	}
	prePayParam.Amount.Total = order.PayMoney
	prePayParam.Amount.Currency = "CNY"
	prePayParam.Payer.OpenId = userOpenId
	reqBody, _ := json.Marshal(prePayParam)
	token, err := wpl.getToken(http.MethodPost, string(reqBody), prePayApiUrl)
	if err != nil {
		err = errcode.Wrap("WxPayLibCreatePrePayError", err)
		return
	}
	_, replyBody, err := httptool.Post(wpl.ctx, prePayApiUrl, reqBody, httptool.WithHeaders(map[string]string{
		"Authorization": "WECHATPAY2-SHA256-RSAa2048" + token,
	}))
	if err != nil {
		err = errcode.Wrap("WxPayLibCreatePrePayError", err)
		return
	}
	prepayReply := struct {
		PrePayId string `json:"pre_pay_id"`
	}{}
	if err = json.Unmarshal(replyBody, &prepayReply); err != nil {
		err = errcode.Wrap("WxPayLibCreatePrePayError", err)
		return
	}
	payInvokeInfo, err = wpl.genPayInvokeInfo(prepayReply.PrePayId)
	if err != nil {
		err = errcode.Wrap("WxPayLibCreatePrePayError", err)
	}
	return payInvokeInfo, nil
}

func (wpl *WxPayLib) getToken(httMethod string, requestBody string, wxApiUrl string) (token string, err error) {

	urlPart, err := url.Parse(wxApiUrl)
	if err != nil {
		return token, err
	}
	canonicalUrl := urlPart.RequestURI()
	timestamp := time.Now().Unix()
	nonce := util.RandomString(32)
	message := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", httMethod, canonicalUrl, timestamp, nonce, requestBody)
	// 商户私有证书放在了 resources 目录下
	pemFileReader, err := resources.LoadResourceFile("wxpay.private.pem")
	if err != nil {
		return token, err
	}
	privateKey, err := ioutil.ReadAll(pemFileReader)
	if err != nil {
		return token, err
	}

	sha256MsgBytes := util.SHA256HashBytes(message)
	signBytes, err := util.RsaSignPKCS1v15(sha256MsgBytes, privateKey, crypto.SHA256)
	if err != nil {
		return token, err
	}
	sign := base64.StdEncoding.EncodeToString(signBytes)

	token = fmt.Sprintf("mchid=\"%s\",nonce_str=\"%s\",timestamp=\"%d\",serial_no=\"%s\",signature=\"%s\"",
		wpl.payConfig.MchId, nonce, timestamp, wpl.payConfig.PrivateSerialNo, sign)
	return token, nil
}

func (wpl *WxPayLib) genPayInvokeInfo(prePayId string) (payInvokeInfo *WxPayInvokeInfo, err error) {
	payInvokeInfo = &WxPayInvokeInfo{
		AppId:     wpl.payConfig.AppId,
		TimeStamp: fmt.Sprintf("%v", time.Now().Unix()),
		NonceStr:  util.RandomString(32),
		Package:   "prepay_id=" + prePayId,
		SignType:  "RSA",
	}
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n", payInvokeInfo.AppId, payInvokeInfo.TimeStamp, payInvokeInfo.NonceStr, payInvokeInfo.Package)
	pemFileReader, err := resources.LoadResourceFile("wxpay.private.pem")
	if err != nil {
		return
	}
	privateKey, err := ioutil.ReadAll(pemFileReader)
	if err != nil {
		return
	}
	sha256MsgBytes := util.SHA256HashBytes(message)
	signBytes, err := util.RsaSignPKCS1v15(sha256MsgBytes, privateKey, crypto.SHA256)
	if err != nil {
		return
	}
	payInvokeInfo.PaySign = base64.StdEncoding.EncodeToString(signBytes)
	return payInvokeInfo, nil
}

func (wpl *WxPayLib) ValidateNotifySingature(timeStamp, nonce, signature, rawPost string) (verifyRes bool, err error) {
	signatureBytes, _ := base64.StdEncoding.DecodeString(signature)
	message := fmt.Sprintf("%s\n%s\n%s\n", timeStamp, nonce, rawPost)
	pemFileReader, err := resources.LoadResourceFile("wxp_pub.pem")
	if err != nil {
		err = errcode.Wrap("WxPayLibValidateCallBackSignatureError", err)
		return
	}
	publicKeyStr, _ := ioutil.ReadAll(pemFileReader)
	block, _ := pem.Decode(publicKeyStr)
	publicKeyInterface, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	publicKey := publicKeyInterface.PublicKey.(*rsa.PublicKey)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, util.SHA256HashBytes(message), signatureBytes)
	verifyRes = nil == err
	return verifyRes, err
}

func (wpl *WxPayLib) DecryptNotifyResourceData(rawPost string) (notifyResourceData *WxPayNotifyResourceData, err error) {
	var notifyResponse WxPayNotifyResponse
	if err = json.Unmarshal([]byte(rawPost), &notifyResponse); nil != err {
		return notifyResourceData, errcode.Wrap("WxPayLibDecryptNotifyDataError", err)
	}
	aseKey := []byte(wpl.payConfig.AesKey)
	nonce := []byte(notifyResponse.Resource.Nonce)
	associatedData := []byte(notifyResponse.Resource.AssociatedData)
	ciphertext, err := base64.StdEncoding.DecodeString(notifyResponse.Resource.Ciphertext)
	if err != nil {
		return notifyResourceData, errcode.Wrap("WxPayLibDecryptNotifyDataError", err)
	}
	block, _ := aes.NewCipher(aseKey)
	aesGcm, _ := cipher.NewGCM(block)
	plaintext, err := aesGcm.Open(nil, nonce, ciphertext, associatedData)
	if err != nil {
		return notifyResourceData, err
	}
	err = json.Unmarshal(plaintext, &notifyResourceData)
	return
}
