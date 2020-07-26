package sqlite3client

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_sqlite3Client_QueryCallback(t *testing.T) {
	dirExists := func(filename string) bool {
		info, err := os.Stat(filename)
		if os.IsNotExist(err) {
			return false
		}
		return info.IsDir()
	}

	if !dirExists("/tmp") {
		_ = os.Mkdir("/tmp", os.ModeDir)
	}
	s, err := New("file:/tmp/test.db")
	defer func() {
		_ = s.Close()
		_ = os.Remove("/tmp/test.db")
	}()
	fmt.Println(err)
	_, err = s.Exec(`
	CREATE TABLE test (
		id integer NOT NULL PRIMARY KEY AUTOINCREMENT
	 ,  name varchar(300) NOT NULL
	 ,  UNIQUE (id)
	 );`)
	fmt.Println(err)
	cont := 0

	fn := func(columns []string, row []interface{}) error {
		fmt.Println(columns, row)
		cont++
		return nil
	}

	for i := 0; i < 10; i++ {
		_, err = s.Exec("INSERT INTO test VALUES(?,?);", i, fmt.Sprintf("test%d", i))
		fmt.Println(err)
	}
	type args struct {
		q    string
		fn   SqlFN
		args []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    int
	}{
		{
			name: "Count",
			args: args{
				q:  "SELECT count(*) FROM test;",
				fn: fn,
			},
			wantErr: false,
			want:    1,
		},
		{
			name: "All",
			args: args{
				q:  "SELECT * FROM test;",
				fn: fn,
			},
			wantErr: false,
			want:    10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cont = 0
			if err := s.QueryCallback(tt.args.fn, tt.args.q, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("sqlite3Client.QueryCallback() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cont != tt.want {
				t.Errorf("sqlite3Client.QueryCallback() error = %v, wantErr %v", cont, tt.want)
			}
		})
	}
}
