package utils

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"github.com/go-faker/faker/v4"
)

func init(){
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomInt(min, max int64) int64 {
	// in fact return number between [min - max +1]
	return min + rand.Int63n(max - min + 1)
}

type RandomAccountParams struct {
	Owner string `faker:"name"`
	Balance int64
	Currency string `faker:"oneof: USR, EUR"`
}

func RandomAccount() RandomAccountParams{
	ra := RandomAccountParams{}
	err := faker.FakeData(&ra)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ra)
	return ra
}

type RandomUserParams struct {
	Username string `faker:"username"`
	HashedPassword string `faker:"password"`
	FullName string `faker:"name"`
	Email string `faker:"email"`
}


func RandomUser() RandomUserParams{
	ru := RandomUserParams{}
	err := faker.FakeData(&ru)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ru)
	return ru
}