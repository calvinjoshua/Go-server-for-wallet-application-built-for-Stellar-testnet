package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
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

	source := os.Getenv("SEED1")

	client := horizonclient.DefaultPublicNetClient

	// Load the source account
	sourceKP := keypair.MustParseFull(source)
	sourceAccountRequest := horizonclient.AccountRequest{AccountID: sourceKP.Address()}
	sourceAccount, err := client.AccountDetail(sourceAccountRequest)
	if err != nil {
		log.Panic("Source Acc:  ", err)
	}

	// Build transaction
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			BaseFee:              txnbuild.MinBaseFee,
			Timebounds:           txnbuild.NewInfiniteTimeout(),
			Operations: []txnbuild.Operation{
				&txnbuild.CreateAccount{
					Destination: kp1.Address(),
					Amount:      "0.00001",
				},
			},
		},
	)

	if err != nil {
		log.Fatal("c1", err)
	}

	tx, err = tx.Sign(network.PublicNetworkPassphrase, sourceKP)
	if err != nil {
		log.Fatal("c2", err)
	}

	resp, err := horizonclient.DefaultPublicNetClient.SubmitTransaction(tx)
	if err != nil {
		log.Fatal("c3", err)
	}

	log.Println("Hash:", resp.Hash)
	return kp1
}

func transfer(dest string, amt string, signature string) bool {

	key := "9ddc81b978ae0aac1004044fec15ed7b5b7fe1f3349ca2365ff65d82e0d0855d"
	diamSeedEncrypted := decrypt(signature, key)
	source := diamSeedEncrypted //os.Getenv("SEED1") //seed goes here
	destination := dest
	client := horizonclient.DefaultPublicNetClient
	log.Println(destination)
	// Make sure destination account exists
	destAccountRequest := horizonclient.AccountRequest{AccountID: destination}
	destinationAccount, err := client.AccountDetail(destAccountRequest)
	if err != nil {
		log.Panic("Destination Acc:  ", err)
	}
	log.Println(destinationAccount)

	// Load the source account
	sourceKP := keypair.MustParseFull(source)
	sourceAccountRequest := horizonclient.AccountRequest{AccountID: sourceKP.Address()}
	sourceAccount, err := client.AccountDetail(sourceAccountRequest)
	if err != nil {
		log.Panic("Source Acc:  ", err)
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
		return false //panic(err)
	}

	tx, err = tx.Sign(network.PublicNetworkPassphrase, sourceKP)
	if err != nil {
		return false //panic(err)
	}

	resp, err := horizonclient.DefaultPublicNetClient.SubmitTransaction(tx)
	if err != nil {
		return false //panic(err)
	}

	fmt.Println(resp)

	return true

}

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

func get(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"Running": "true"})
}

func main() {
	app := fiber.New()

	app.Get("/get", get)

	app.Post("/login", login)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte("secret"), //tokenKey
	}))

	app.Post("/getToken", Regenerate)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
	}))

	app.Get("/check/:add", func(c *fiber.Ctx) error {
		var address = c.Params("add")
		var res = check(address)

		return c.JSON(res)
	})
	app.Get("/env", func(c *fiber.Ctx) error {
		return c.SendString("Address: " + os.Getenv("VARNAME"))
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                             Account creation                                                                                                                     //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Get("/AccountCreation", func(c *fiber.Ctx) error {

		var sa = keyPair()
		var seed = sa.Seed()
		var Address = sa.Address()
		return c.JSON(fiber.Map{"Address": Address, "Seed": seed})
		//return c.SendString("Address: " + Address + "\nSeed: " + seed)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////                                                                                                                                                                  //////
	/////                                Native Diam Transfer                                                                                                              //////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	app.Post("/Transfer", func(c *fiber.Ctx) error {
		payload := struct {
			Address   string `json:"address"`
			Amount    string `json:"amount"`
			Signature string `json:"signature"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		address := payload.Address
		amount := payload.Amount
		signature := payload.Signature

		// var sa = c.Params("address")
		// var amt = c.Params("amt")

		var val = check(address)
		if val == "valid" {
			resp := transfer(address, amount, signature)
			if resp {
				return c.JSON(fiber.Map{"statusCode": 200, "message": "Transferred"})
			} else {
				return c.JSON(fiber.Map{"statusCode": 500, "message": "Transfer failed, internal server error"})
			}
		} else {
			return c.JSON(fiber.Map{"statusCode": 401, "message": "INVLAID ADDRESS"})
		}
	})

	app.Listen(":3000")
}
