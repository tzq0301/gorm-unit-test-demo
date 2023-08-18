package user

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"gorm-unit-test-demo/repo/test"
	"gorm-unit-test-demo/repo/user"
)

func TestRepo_Get(t *testing.T) {
	a := require.New(t)

	db, postSetup := test.SetupLocalDatabaseConnection(t, user.User{}.TableName())
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
		want    user.User
		wantErr bool
	}{
		{"Got", fields{db}, args{1}, user.User{1, "Tony", 18}, false},
		{"Err", fields{db}, args{4}, user.User{0, "", 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &user.Repo{
				DB: tt.fields.db,
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
