/*
This is a very simple package that implements
admin entity for our system. But this package should be removed
in the future and it has to be replaced by the Fabric's admin entity.
*/
package admin

import (
	"../crypdata"
	"../onchain"
	"errors"
)

// This is so simple admin entity struct
// Ideally, we have to use admin entity struct in the Fabric
type Admin struct {
	Hashedpassword string
}

var mainAdmin *Admin = nil

// Init() makes mainAdmin var and enroll admin entity of the Fabric
func Init(adminPassw string) error {
	if mainAdmin != nil {
		return errors.New("admin already exists")
	}
	mainAdmin = new(Admin)
	mainAdmin.Hashedpassword = crypdata.Hash(adminPassw)
	err := onchain.EnrollAdmin()
	if err != nil {
		mainAdmin = nil
		return errors.New("Failed enroll admin")
	}
	return nil
}

// IsAdminPassword() checks if passw matches admin password
// Here there may be a digital sign verification
func IsAdminPassword(passw string) bool {
	return crypdata.Hash(passw) == mainAdmin.Hashedpassword
}
