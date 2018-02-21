/*
This package implements a connection between the offchain and onchain parts.
Its main purpose is to provide maintenance of nodejs sdk
(launching js scripts in a separate processes)

There are a lot of duplicate stupid code here, but this code
should be replaced by go sdk layer.
*/
package onchain

import (
	"bytes"
	"errors"
	"os/exec"
	"regexp"
)

func EnrollAdmin() error {
	outCmd := exec.Command("node", "../fabusers/enrollAdmin.js")

	var out bytes.Buffer
	outCmd.Stdout = &out
	err := outCmd.Run()
	return err
}

func RegisterUser(username *string) error {
	outCmd := exec.Command("node", "../fabusers/registerUser.js", *username)

	var out bytes.Buffer
	outCmd.Stdout = &out
	err := outCmd.Run()
	return err
}

func GetUserhash(username *string) (string, error) {
	outCmd := exec.Command("node", "../fabusers/query.js", *username)
	var out bytes.Buffer
	outCmd.Stdout = &out
	err := outCmd.Run()
	if err != nil {
		return "", err
	}
	output := out.String()
	re := regexp.MustCompile("OK RESPONSE: \\{\"info_hash\":\"(.*?)\"\\}")
	match := re.FindStringSubmatch(output)

	if len(match) == 0 {
		return "", errors.New("Check username")
	} else {
		return match[1], nil
	}
}

func AddUserInfoToLedger(username *string, userhash *string) error {
	outCmd := exec.Command("node", "../fabusers/addUser.js", *username, *userhash)
	var out bytes.Buffer
	outCmd.Stdout = &out
	err := outCmd.Run()
	return err
}

func UpdateLedgerUserinfo(username *string, userhash *string) error {
	outCmd := exec.Command("node", "../fabusers/changeUserInfoHash.js", *username, *userhash)
	var out bytes.Buffer
	outCmd.Stdout = &out
	err := outCmd.Run()
	return err
}
