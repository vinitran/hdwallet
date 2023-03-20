package databases

import "testing"

func TestConnectDatabase(t *testing.T) {
	_, err := ConnectDatabase()
	if err != nil {
		t.Error(err)
	}
}
