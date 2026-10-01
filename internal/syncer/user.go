package syncer

import (
	"errors"
	"strconv"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type User struct {
	Username     string
	RealName     string
	Email        string
	BuildingID   string
	DepartmentID string
}

var buildingIDs = map[string]string{
	"No Building": "-1", "Senior Campus": "1", "Penbank": "2", "Minimbah": "3", "Disabled": "4",
}

var departmentIDs = map[string]string{
	"No Department": "-1", "Y7": "5", "Y8": "6", "Y9": "7", "Y10": "8", "Y11": "9", "Y12": "10",
	"Staff": "11", "Y6": "12", "Y5": "13", "Y4": "14", "Y3": "15", "Y2": "16", "Y1": "17",
	"Y0": "18", "ELC": "19", "Y13": "20", "IT": "21", "Testing": "22", "IWBs": "23", "Spares": "24", "NA": "25", "Disabled": "26",
}

func directoryUser(entry *ldap.Entry) (User, error) {
	username := entry.GetEqualFoldAttributeValue("sAMAccountName")
	control, err := strconv.ParseUint(entry.GetEqualFoldAttributeValue("userAccountControl"), 10, 32)
	if err != nil || username == "" {
		return User{}, errors.New("directory: required account attributes are missing or invalid")
	}
	user := User{
		Username: strings.ToLower(username), RealName: entry.GetEqualFoldAttributeValue("name"), Email: entry.GetEqualFoldAttributeValue("mail"),
		BuildingID:   mappedID(buildingIDs, entry.GetEqualFoldAttributeValue("Campus")),
		DepartmentID: mappedID(departmentIDs, entry.GetEqualFoldAttributeValue("department")),
	}
	if control&0x2 != 0 {
		user.BuildingID, user.DepartmentID = "4", "26"
	}
	return user, nil
}

func mappedID(ids map[string]string, value string) string {
	if id, ok := ids[value]; ok {
		return id
	}
	return "-1"
}
