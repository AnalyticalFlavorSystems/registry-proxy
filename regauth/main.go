package main

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/mewbak/gopass"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "regauth"
	app.Usage = "add new users to registry"
	app.Commands = []cli.Command{
		{
			Name:   "add",
			Usage:  "add a new user",
			Action: addUser,
		},
	}
	app.Run(os.Args)
}
func addUser(c *cli.Context) {
	if _, err := os.Stat("registry.db"); os.IsNotExist(err) {
		log.Fatal("cant't find registry.db")
	}

	user := c.Args().First()

	// Password
	pwd, err := gopass.GetPass("Password:")
	if err != nil {
		log.Fatal(err)
	}
	pwdByte := []byte(pwd)

	// Password Match
	pwdMatch, err := gopass.GetPass("Confirm Password:")
	if err != nil {
		log.Fatal(err)
	}
	pwdMatchByte := []byte(pwdMatch)

	if !bytes.Equal(pwdByte, pwdMatchByte) {
		log.Fatal("Passwords don't match")
	}
	db, err := bolt.Open("registry.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	password := SetPassword(pwd)

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("auth"))
		if err != nil {
			return err
		}
		err = b.Put([]byte(user), password)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successful")
}
func SetPassword(password string) []byte {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hpass
}
