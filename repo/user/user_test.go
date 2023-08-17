package user

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"gorm-unit-test-demo/repo/test"
)

//go:embed user.sql
var userTableDDL string

func TestRepo_Get(t *testing.T) {
	a := require.New(t)

	db, postSetup := test.SetupLocalDatabaseConnection(t, User{}.TableName(), userTableDDL)
	defer postSetup()

	a.NoError(db.Exec("INSERT INTO `user`(`name`, `age`) VALUES ('Tony', 18)").Error)
	a.NoError(db.Exec("INSERT INTO `user`(`name`, `age`) VALUES ('White', 28)").Error)
	a.NoError(db.Exec("INSERT INTO `user`(`name`, `age`) VALUES ('Milo', 12)").Error)

	type fields struct {
		db *gorm.DB
	}
	type args struct {
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    User
		wantErr bool
	}{
		{"Got", fields{db}, args{1}, User{1, "Tony", 18}, false},
		{"Err", fields{db}, args{4}, User{0, "", 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repo{
				db: tt.fields.db,
			}
			got, err := r.Get(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}
