/*
This is an onchain layer

TODO: description
*/
package onchain

import (
	"errors"
	"bytes"
	"os/exec"
	"regexp"
)

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