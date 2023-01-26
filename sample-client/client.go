package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

func main() {
	//セッションの確立
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	cognitoClient := cognitoidentityprovider.New(sess)

	userpoolID := flag.String("i", "", "userpoolID")
	userName := flag.String("u", "", "username")
	password := flag.String("p", "", "password")
	flag.Parse()

	if *userpoolID == "" || *userName == "" || *password == "" {
		fmt.Println("poolID, username or password is empty")
		return
	}

	newUserData := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        userpoolID,
		Username:          userName,
		TemporaryPassword: password,
	}

	_, err := cognitoClient.AdminCreateUser(newUserData)
	if err != nil {
		fmt.Println(err)
		return
	}

}
