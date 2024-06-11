package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/Xiippy/Xiippy.POSeCommSDK.Light_GoLang/Models"
	"github.com/Xiippy/Xiippy.POSeCommSDK.Light_GoLang/XiippySDKBridgeApiClient"
	"github.com/google/uuid"
)

// page data model
type PageData struct {
	ErrorText      string
	XiippyFrameUrl string
}

// the main handler of the payment page
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	// try initiating the payment and loading the payment card screen
	XiippyFrameUrl, err := InitiatePaymentNGetiFrameUrl()

	data := PageData{
		XiippyFrameUrl: XiippyFrameUrl,
	}

	if err != nil {
		// show the error message, if any
		data.ErrorText = err.Error()
	}

	tmpl, err := template.ParseFiles("templates/layout.html", "templates/home.html")
	if err != nil {
		// show the error message, if any
		data.ErrorText = err.Error()
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func BuildQueryString(keyValuePairs map[string]string) string {
	queryString := url.Values{}

	for key, value := range keyValuePairs {
		queryString.Add(key, value)
	}

	return queryString.Encode()
}

type Config struct {
	Config_BaseAddress  string `json:"Config_BaseAddress"`
	Config_ApiSecretKey string `json:"Config_ApiSecretKey"`
	MerchantID          string `json:"MerchantID"`
	MerchantGroupID     string `json:"MerchantGroupID"`
}

func InitiatePaymentNGetiFrameUrl() (string, error) {

	// you would load these from secure config management for each merchant (e.g. the merchant relevant to a selected store on an e-commerce site)
	// a merchant group represents a franchise or a group of logically-connected merchants
	// another way to look at a merchant group it is like a platform when processing payments on behalf of merchants

	config := Config{
		Config_BaseAddress:  "The base URL of the SDK Bridge instance",
		MerchantGroupID:     "Your Merchant Group ID, as reported by the SDK Bridge instance",
		Config_ApiSecretKey: "API Secret Key, as reported by the SDK Bridge instance",
		MerchantID:          "Your Merchant ID, as reported by the SDK Bridge instance",
	}

	// depending on the basket, shipping and billing address entered, as well as amounts, the payment is initialized:
	StatementID := uuid.New().String()
	UniqueStatementID := uuid.New().String()
	req := Models.PaymentProcessingRequest{
		MerchantGroupID:  config.MerchantGroupID,
		MerchantID:       config.MerchantID,
		Amount:           2.5,
		Currency:         "aud",
		ExternalUniqueID: UniqueStatementID,
		IsPreAuth:        false,
		IsViaTerminal:    false,
		// customer is optional
		Customer: &Models.PaymentRecordCustomer{
			CustomerAddress: Models.PaymentRecordCustomerAddress{
				CityOrSuburb:    "Brisbane",
				Country:         "Australia",
				FullName:        "Full Name",
				Line1:           "100 Queen St",
				PhoneNumber:     "+61400000000",
				PostalCode:      "4000",
				StateOrPrivince: "Qld",
			},
			CustomerEmail: "dont@contact.me",
			CustomerName:  "Full Name",
			CustomerPhone: "+61400000000",
		},
		IssuerStatementRecord: &Models.IssuerStatementRecord{
			// this could be a different id than RandomStatementID which is a Xiippy identifier
			UniqueStatementID:        UniqueStatementID,
			RandomStatementID:        StatementID,
			StatementCreationDate:    time.Now().UTC().String(),
			StatementTimeStamp:       time.Now().Format("20060102150405"),
			Description:              "Test transaction #1",
			DetailsInBodyBeforeItems: "Description on the receipt before items",
			DetailsInBodyAfterItems:  "Description on the receipt after items",
			DetailsInFooter:          "Description on the footer",
			DetailsInHeader:          "Description on the header",
			StatementItems: []Models.StatementItem{
				{
					Description: "Description",
					UnitPrice:   11,
					Url:         "Url",
					Quantity:    1,
					Identifier:  "Identifier",
					Tax:         1,
					TotalPrice:  11,
				},
				{
					Description: "Description2",
					UnitPrice:   33,
					Url:         "Url2",
					Quantity:    1,
					Identifier:  "Identifier2",
					Tax:         3,
					TotalPrice:  33,
				},
			},
			TotalAmount:    44,
			TotalTaxAmount: 4,
		},
	}

	// instantiate the sdk client and pass the parameters
	client := XiippySDKBridgeApiClient.NewXiippySDKBridgeApiClient(true, config.Config_ApiSecretKey, config.Config_BaseAddress, config.MerchantID, config.MerchantGroupID)

	// initiate the payment
	response, err := client.InitiateXiippyPayment(&req)
	if err != nil {
		return "", err
	}

	// compile the parameters required to load and initialize the payments screen
	QueryString := BuildQueryString(map[string]string{
		XiippySDKBridgeApiClient.QueryStringParamRsid:               response.RandomStatementID,
		XiippySDKBridgeApiClient.QueryStringParamSts:                response.StatementTimeStamp,
		XiippySDKBridgeApiClient.QueryStringParamCa:                 response.ClientAuthenticator,
		XiippySDKBridgeApiClient.QueryStringParamSpw:                "true", // show plain view
		XiippySDKBridgeApiClient.QueryStringParamMerchantID:         config.MerchantID,
		XiippySDKBridgeApiClient.QueryStringParamMerchantGroupID:    config.MerchantGroupID, // important
		XiippySDKBridgeApiClient.QueryStringParamCs:                 response.ClientSecret,
		XiippySDKBridgeApiClient.QueryStringParamShowLongXiippyText: "true", // show the long xiippy description text
	})

	FullPaymentPageUrl := fmt.Sprintf("%s/Payments/Process?%s", config.Config_BaseAddress, QueryString)
	// optional
	fmt.Printf("The payment page can not be browsed at '%s'\n", FullPaymentPageUrl)

	return FullPaymentPageUrl, nil
}
