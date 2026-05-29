package a

import "testing"

type fakeDB struct{}

func (fakeDB) Exec(query string, args ...interface{}) (interface{}, error) {
	return nil, nil
}
func (fakeDB) Query(query string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

var db = fakeDB{}

func TestBad(t *testing.T) {
	db.Exec("INSERT INTO users VALUES (1)")        // want `direct DB mutation in test`
	db.Exec("DELETE FROM users WHERE id = 1")      // want `direct DB mutation in test`
	db.Exec("UPDATE users SET name='x' WHERE id=1") // want `direct DB mutation in test`
	db.Query("TRUNCATE TABLE users")               // want `direct DB mutation in test`
	db.Query("SELECT * FROM users")
}
