package main

import (
	"log"
	"os"

	//"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
	//"github.com/golang-jwt/jwt/v4"
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

func transfer(dest string, amt string) {
	source := os.Getenv("SEED1")
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
		panic(err)
	}

	tx, err = tx.Sign(network.PublicNetworkPassphrase, sourceKP)
	if err != nil {
		panic(err)
	}

	resp, err := horizonclient.DefaultPublicNetClient.SubmitTransaction(tx)
	if err != nil {
		panic(err)
	}
	log.Println("Hash:", resp.Hash)

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
	app.Post("/Transfer/:address/:amt", func(c *fiber.Ctx) error {

		var sa = c.Params("address")
		var amt = c.Params("amt")

		var val = check(sa)
		if val == "valid" {
			transfer(sa, amt)
			return c.SendString("Credited to Address: " + sa)
		} else {
			return c.SendString("INVALID ADDRESS: " + sa)
		}
	})

	app.Listen(":3000")
}
