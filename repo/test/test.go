package test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	testUseLocalDatabaseMySQLForDaoUnitTestKey   = "USE_LOCAL_DATABASE_MYSQL_FOR_DAO_UNIT_TEST"
	testUseLocalDatabaseMySQLForDaoUnitTestValue = "1"
)

// SetupLocalDatabaseConnection 为 DAO 层的单元测试创建本地 MySQL 连接，并提供自动删表的方法
//
// 需要设置环境变量 USE_LOCAL_DATABASE_MYSQL_FOR_DAO_UNIT_TEST=1（否则单元测试方法将被 Skip）
//
// For example, create the table `user_info` and drop the table automatically at the end of function by defer
//
//	func TestSomething(t *testing.T) {
//	    a := assert.New(T)
//
//	    db, postSetup := test.SetupDatabase(t, "user_info", "CREATE TABLE `user_info` ( ... );")
//	    defer postSetup()
//
//	    // Insert 100 records
//
//	    var count int64
//	    db.Model(&UserInfo{}).Count(&count)
//	    a.EqualValues(100, count)
//
//	    // At the end of function, the function `postSetup` will be called, and it will drop the table `user_info` if `defer` is used
//	}
//
// Recommend use go:embed to read DDL from SQL file
func SetupLocalDatabaseConnection(t *testing.T, tableName, ddl string) (db *gorm.DB, postSetup func()) {
	if !testing.Testing() {
		log.Fatalln("don't call this function in non-test mode")
	}

	if t == nil {
		log.Fatalf("%q is nil\n", "t *testing.T")
	}

	if os.Getenv(testUseLocalDatabaseMySQLForDaoUnitTestKey) != testUseLocalDatabaseMySQLForDaoUnitTestValue {
		t.SkipNow()
	}

	// How to set up a MySQL Instance quickly by Docker? Refer to https://zhuanlan.zhihu.com/p/635819648
	// 1. docker pull mysql
	// 2. docker run -p 3306:3306 --name mysql -e MYSQL_ROOT_PASSWORD=123456 -d mysql
	// 3. docker exec -it mysql bash
	// 4. mysql -uroot -p123456
	// 5. CREATE DATABASE `test`;

	config := struct {
		host     string
		port     int
		username string
		password string
		database string
	}{
		host:     "127.0.0.1",
		port:     3306,
		username: "root",
		password: "123456",
		database: "test",
	}

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		config.username, config.password, config.host, config.port, config.database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		t.Fatal("can't connect to database")
		return
	}

	db.Exec(ddl)

	return db, func() {
		err = db.Migrator().DropTable(tableName)
		if err != nil {
			t.Fatalf("can't drop table %v", tableName)
		}
	}
}
