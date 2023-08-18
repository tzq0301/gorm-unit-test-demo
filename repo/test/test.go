package test

import (
	"embed"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	testUseLocalDatabaseMySQLForDaoUnitTestKey   = "USE_LOCAL_DATABASE_MYSQL_FOR_DAO_UNIT_TEST"
	testUseLocalDatabaseMySQLForDaoUnitTestValue = "1"

	sqlRoot = "sql"
)

var (
	once sync.Once
	db   *gorm.DB

	//go:embed sql
	sqlFS embed.FS
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
//	    db, postSetup := test.SetupDatabase(t, "user_info", "user_register")
//	    defer postSetup()
//
//	    // Insert 100 records
//
//	    var count int64
//	    db.Model(&UserInfo{}).Count(&count)
//	    a.EqualValues(100, count)
//
//	    // At the end of function, the function `postSetup` will be called, and it will truncate the table `user_info` and `user_register` if `defer` is used
//	}
func SetupLocalDatabaseConnection(t *testing.T, tableNames ...string) (*gorm.DB, func()) {
	// Go 1.21 enabled, if your go version is below 1.21, delete this block
	if !testing.Testing() {
		log.Fatalln("don't call this function in non-test mode")
	}

	if t == nil {
		log.Fatalf("%q is nil\n", "t *testing.T")
	}

	if os.Getenv(testUseLocalDatabaseMySQLForDaoUnitTestKey) != testUseLocalDatabaseMySQLForDaoUnitTestValue {
		t.SkipNow()
	}

	if len(tableNames) == 0 {
		t.Fatal("empty table name list")
	}

	for _, tableName := range tableNames {
		if len(strings.TrimSpace(tableName)) == 0 {
			t.Fatal("empty table name")
		}
	}

	// How to set up a MySQL Instance quickly by Docker? Refer to https://zhuanlan.zhihu.com/p/635819648
	// 1. docker pull mysql
	// 2. docker run -p 3306:3306 --name mysql -e MYSQL_ROOT_PASSWORD=123456 -d mysql
	// 3. docker exec -it mysql bash
	// 4. mysql -uroot -p123456
	// 5. CREATE DATABASE `test`;

	once.Do(func() {
		// 1. Connect to MySQL
		// 2. Create Table

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
		initDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			t.Fatal("can't connect to database")
		}

		db = initDB

		entries, err := sqlFS.ReadDir(sqlRoot)
		if err != nil {
			t.Fatal("can't get SQL files")
		}

		for _, entry := range entries {
			sqlFilePath := fmt.Sprintf("%s/%s", sqlRoot, entry.Name())

			bytes, err := sqlFS.ReadFile(sqlFilePath)
			if err != nil {
				t.Fatalf("can't get SQL file: %v", entry.Name())
			}

			sqls := strings.Split(string(bytes), ";\n")

			for _, sql := range sqls {
				if len(strings.TrimSpace(strings.ReplaceAll(sql, "\n", ""))) == 0 {
					continue
				}

				if db.Exec(sql).Error != nil {
					t.Fatalf("can't exec sql in file: %s", sqlFilePath)
				}
			}
		}
	})

	session := db.Session(&gorm.Session{})

	sort.Strings(tableNames)

	if session.Exec(fmt.Sprintf("LOCK TABLES %s WRITE", strings.Join(tableNames, " WRITE, "))).Error != nil {
		t.Fatalf("can't lock tables %v", tableNames)
	}

	postSetup := func() {
		for _, tableName := range tableNames {
			if session.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", tableName)).Error != nil {
				t.Fatalf("can't drop table %v", tableName)
			}
		}

		if session.Exec("UNLOCK TABLES").Error != nil {
			t.Fatalf("can't unlock tables")
		}
	}

	return session, postSetup
}
