/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////                                                                                                                                                                  //////
/////                     Only Account creation and transfer are major required API's, rest can be used by using the endpoint https://diamtestnet.diamcircle.io/       //////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

//The below function creates a keypair and funds that address to make it an account
func keyPair() *keypair.Full {
	kp1, err := keypair.Random()
	if err != nil {
		log.Fatal(err)
	}
	source := "SCP4WMISYL3TQRG6AQSV2CMUOSEWXPZ6WDOBNXOHKHMNIIYH2SHPGK2L"

	client := horizonclient.DefaultTestNetClient

	// Load the source account
	sourceKP := keypair.MustParseFull(source)
	sourceAccountRequest := horizonclient.AccountRequest{AccountID: sourceKP.Address()}
	sourceAccount, err := client.AccountDetail(sourceAccountRequest)
	if err != nil {
		panic(err)
	}

	// Build transaction
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			BaseFee:              txnbuild.MinBaseFee,
			Timebounds:           txnbuild.NewInfiniteTimeout(), // Use a real timeout in production!
			Operations: []txnbuild.Operation{
				&txnbuild.CreateAccount{ //operation
					Destination: kp1.Address(),
					Amount:      "1", //change the diam quantity if neede a change
				},
			},
		},
	)

	if err != nil {
		panic(err)
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, sourceKP)
	if err != nil {
		panic(err)
	}

	resp, err := horizonclient.DefaultTestNetClient.SubmitTransaction(tx)
	if err != nil {
		panic(err)
	}

	log.Println("Hash:", resp.Hash)
	return kp1
}

//Currently only from know accounts
func transfer(dest string, amt string) {
	//kp, _ := keypair.Parse("SCZANGBA5YHTNYVVV4C3U252E2B6P6F5T3U6MM63WBSBZATAQI3EBTQ4")
	source := "SDUANIURZ7B6MH4DXGGTITXFPY7ISM6MZ3HE7NC6CSA7D5LA44HLDMHX"
	destination := dest
	client := horizonclient.DefaultTestNetClient
	log.Println(destination)
	// Make sure destination account exists
	destAccountRequest := horizonclient.AccountRequest{AccountID: destination}
	log.Println("p1")
	destinationAccount, err := client.AccountDetail(destAccountRequest)
	if err != nil {
		log.Panic("caught ", err)
	}
	log.Println(destinationAccount)

	// Load the source account
	sourceKP := keypair.MustParseFull(source)
	sourceAccountRequest := horizonclient.AccountRequest{AccountID: sourceKP.Address()}
	sourceAccount, err := client.AccountDetail(sourceAccountRequest)
	if err != nil {
		panic(err)
	}

	// Build transaction
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			BaseFee:              txnbuild.MinBaseFee,
			Timebounds:           txnbuild.NewInfiniteTimeout(), // Use a real timeout in production!
			Operations: []txnbuild.Operation{
				&txnbuild.Payment{
					Destination: destination,
					Amount:      amt,
					Asset:       txnbuild.NativeAsset{},
				},
			},
		},
	)

	if err != nil {
		panic(err)
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, sourceKP)
	if err != nil {
		panic(err)
	}

	resp, err := horizonclient.DefaultTestNetClient.SubmitTransaction(tx)
	if err != nil {
		panic(err)
	}
	log.Println("Hash:", resp.Hash)

}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////                                                                                                                                                                  //////
/////                             This implementation is not really needed, for balance, this api can be called https://diamtestnet.diamcircle.io/accounts             //////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func balance(address string) (string, string) {
	type Ping struct {
		Balance             string `json:"balance"`
		Buying_liabilities  string `json:"buying_liabilities"`
		Selling_liabilities string `json:"selling_liabilities"`
		Asset_type          string `json:"asset_type"`
	}
	type PingContent struct {
		Balance []Ping `json:"balances"`
	}
	httpposturl := "http://10.0.34.42:8000/accounts/" + address
	request, error := http.NewRequest("GET", httpposturl, nil)
	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	content := PingContent{}
	json.Unmarshal([]byte(body), &content)
	return content.Balance[0].Asset_type, content.Balance[0].Balance
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////                                                                                                                                                                  //////
/////                             This implementation is to check if an address is valid                                                                               //////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func check(dest string) string {

	client := horizonclient.DefaultTestNetClient
	destAccountRequest := horizonclient.AccountRequest{AccountID: dest}
	destinationAccount, err := client.AccountDetail(destAccountRequest)
	if err != nil {
		return "invalid"
	}
	log.Println(destinationAccount)
	return "valid"

}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////                                                                                                                                                                  //////
/////                             main()                                                                                                                               //////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
	}))

	app.Get("/check/:add", func(c *fiber.Ctx) error {
		var address = c.Params("add")
		var res = check(address)

		return c.JSON(res)
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                             Account creation                                                                                                                     //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/AccountCreation", func(c *fiber.Ctx) error {
		var sa = keyPair()
		var seed = sa.Seed()
		var Address = sa.Address()
		return c.SendString("Address: " + Address + "\nSeed: " + seed)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                                                                                                                                                  //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/balance/:address", func(c *fiber.Ctx) error {
		type details struct {
			AssetName string
			Balance   string
		}
		var address = c.Params("address")
		var at, bl = balance(address)
		content := details{}
		content.AssetName = at
		content.Balance = bl
		return c.JSON(content)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                                                                                                                                                  //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//!!!!!!!! To merge the effects and TransactionAcc, use created_at as reference
	app.Get("/AccountEffects/:address", func(c *fiber.Ctx) error {
		var address = c.Params("address")
		type Ping1 struct {
			Action     string `json:"type"`
			Amount     string `json:"amount"`
			Created_at string `json:"created_at"`
			Asset      string `json:"asset_type"`
		}
		type Ping struct {
			Prop []Ping1 `json:"records"` //Prop Ping1 `json:"self"`
		}
		type PingContent struct {
			Links Ping `json:"_embedded"`
		}
		httpposturl := "http://10.0.34.42:8000/accounts/" + address + "/effects" //"http://10.0.34.42:8000/transaction/cbd20b9f70ec83f49f414deb9ee149acaebdfe6fb9d77f3ab171e8406faeb390"
		request, error := http.NewRequest("GET", httpposturl, nil)
		client := &http.Client{}
		response, error := client.Do(request)
		if error != nil {
			panic(error)
		}
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		var content = new(PingContent) //content := PingContnt{}
		json.Unmarshal([]byte(body), &content)
		return c.JSON(content.Links.Prop)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                Native Diam Transfer                                                                                                              //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Post("/Transfer/:address/:amt", func(c *fiber.Ctx) error {

		var sa = c.Params("address")
		var amt = c.Params("amt")
		transfer(sa, amt)

		return c.SendString("Address: " + sa)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                                                                                                                                                  //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/TransactionByAccount/:address", func(c *fiber.Ctx) error {
		var address = c.Params("address")
		type Ping11 struct {
			ID         string `json:"id"`
			Action     string `json:"source_account"`
			Amount     string `json:"hash"`
			Created_at string `json:"created_at"`
			Asset      string `json:"asset_type"`
		}
		type Ping12 struct {
			Prop []Ping11 `json:"records"` //Prop Ping1 `json:"self"`
		}
		type PingContnt1 struct {
			Links Ping12 `json:"_embedded"`
		}

		httpposturl1 := "http://10.0.44.20:8000/accounts/" + address + "/transactions" //"http://10.0.34.42:8000/transaction/cbd20b9f70ec83f49f414deb9ee149acaebdfe6fb9d77f3ab171e8406faeb390"
		request1, error := http.NewRequest("GET", httpposturl1, nil)
		client1 := &http.Client{}
		response1, error := client1.Do(request1)
		if error != nil {
			panic(error)
		}
		defer response1.Body.Close()
		body1, _ := ioutil.ReadAll(response1.Body)
		var content1 = new(PingContnt1) //content := PingContnt{}
		json.Unmarshal([]byte(body1), &content1)
		return c.JSON(content1.Links.Prop)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                                                                                                                                                  //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/TransactionsAll", func(c *fiber.Ctx) error {

		type Ping11 struct {
			ID              string `json:"id"`
			Action          string `json:"source_account"`
			Amount          string `json:"hash"`
			Created_at      string `json:"created_at"`
			Operation_count int    `json:"operation_count"`
			Ledger          int    `json:"ledger"`
		}
		type Ping12 struct {
			Prop []Ping11 `json:"records"` //Prop Ping1 `json:"self"`
		}
		type PingContnt1 struct {
			Links Ping12 `json:"_embedded"`
		}

		httpposturl1 := "http://10.0.44.20:8000/transactions?order=desc" //"http://10.0.34.42:8000/transaction/cbd20b9f70ec83f49f414deb9ee149acaebdfe6fb9d77f3ab171e8406faeb390"
		request1, error := http.NewRequest("GET", httpposturl1, nil)
		client1 := &http.Client{}
		response1, error := client1.Do(request1)
		if error != nil {
			panic(error)
		}
		defer response1.Body.Close()
		body1, _ := ioutil.ReadAll(response1.Body)
		var content1 = new(PingContnt1) //content := PingContnt{}
		json.Unmarshal([]byte(body1), &content1)
		return c.JSON(content1.Links.Prop)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                                                                                                                                                  //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/TransactionOfHash/:hash", func(c *fiber.Ctx) error {
		var hash = c.Params("hash")
		type PingContnt1 struct {
			Successful      bool   `json:"successful"`
			Hash            string `json:"hash"`
			Ledger          int    `json:"ledger"`
			Created_at      string `json:"created_at"`
			Source_account  string `json:"source_account"`
			Fee_charged     string `json:"fee_charged"`
			Operation_count int    `json:"operation_count"`
		}
		httpposturl1 := "http://10.0.44.20:8000/transactions/" + hash //"http://10.0.34.42:8000/transaction/cbd20b9f70ec83f49f414deb9ee149acaebdfe6fb9d77f3ab171e8406faeb390"
		request1, error := http.NewRequest("GET", httpposturl1, nil)
		client1 := &http.Client{}
		response1, error := client1.Do(request1)
		if error != nil {
			panic(error)
		}
		defer response1.Body.Close()
		body1, _ := ioutil.ReadAll(response1.Body)
		var content1 = new(PingContnt1) //content := PingContnt{}
		json.Unmarshal([]byte(body1), &content1)
		return c.JSON(content1)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                                                                                                                                                  //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/ledgerDetails/:ledger", func(c *fiber.Ctx) error {
		var hash = c.Params("ledger")
		type PingContnt1 struct {
			// Successful      bool   `json:"successful"`
			Hash       string `json:"hash"`
			Phash      string `json:"prev_hash"`
			Created_at string `json:"closed_at"`
			// Oc  string `json:"operation_count"`
			Fee_charged     string `json:"fee_pool"`
			Operation_count int    `json:"operation_count"`
		}

		httpposturl1 := "http://10.0.44.20:8000/ledgers/" + hash //"http://10.0.34.42:8000/transaction/cbd20b9f70ec83f49f414deb9ee149acaebdfe6fb9d77f3ab171e8406faeb390"
		request1, error := http.NewRequest("GET", httpposturl1, nil)
		client1 := &http.Client{}
		response1, error := client1.Do(request1)
		if error != nil {
			panic(error)
		}
		defer response1.Body.Close()
		body1, _ := ioutil.ReadAll(response1.Body)
		var content1 = new(PingContnt1) //content := PingContnt{}
		json.Unmarshal([]byte(body1), &content1)
		return c.JSON(content1)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                                                                                                                                                  //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/AccountDetails/:account", func(c *fiber.Ctx) error {
		var address = c.Params("account")
		type Ping struct {
			Balance             string `json:"balance"`
			Buying_liabilities  string `json:"buying_liabilities"`
			Selling_liabilities string `json:"selling_liabilities"`
			Asset_type          string `json:"asset_type"`
		}
		type PingContent struct {
			Aid     string `json:"account_id"`
			Sq      string `json:"sequence"`
			Lmd     int    `json:"last_modified_ledger"`
			Lmt     string `json:"last_modified_time"`
			Balance []Ping `json:"balances"`
			Pt      string `json:"paging_token"`
		}
		httpposturl := "https://diamtestnet.diamcircle.io/accounts/" + address
		request, error := http.NewRequest("GET", httpposturl, nil)
		client := &http.Client{}
		response, error := client.Do(request)
		if error != nil {
			panic(error)
		}
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		content := PingContent{}
		json.Unmarshal([]byte(body), &content)

		return c.JSON(content)
		//return content.Balance[0].Asset_type, content.Balance[0].Balance
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                                                                                                                                                  //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/operation/:id", func(c *fiber.Ctx) error {
		var id = c.Params("id")
		type Ping1 struct {
			SourceAccount   string `json:"source_account"`
			Type            string `json:"type"`
			Created_at      string `json:"created_at"`
			TransactionHash string `json:"transaction_hash"`
			StartingBalance string `json:"starting_balance"`
			Account         string `json:"account"`
		}
		type Ping struct {
			Prop []Ping1 `json:"records"` //Prop Ping1 `json:"self"`
		}
		type PingContent struct {
			Links Ping `json:"_embedded"`
		}
		httpposturl := "http://10.0.34.42:8000/transactions/" + id + "/operations" //"http://10.0.34.42:8000/transaction/cbd20b9f70ec83f49f414deb9ee149acaebdfe6fb9d77f3ab171e8406faeb390"
		request, error := http.NewRequest("GET", httpposturl, nil)
		client := &http.Client{}
		response, error := client.Do(request)
		if error != nil {
			panic(error)
		}
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		var content = new(PingContent) //content := PingContnt{}
		json.Unmarshal([]byte(body), &content)
		return c.JSON(content.Links.Prop)
	})

	app.Listen(":3006")
}
