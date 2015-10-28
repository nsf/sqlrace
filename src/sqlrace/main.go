package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"strings"
)

var iterNum = flag.Int("n", 1024, "number of iterations per goroutine")
var goNum = flag.Int("g", 4, "number of goroutines")
var method = flag.String("m", "naive", "decrement method: naive, transaction, locked, atomic")

type TestTable struct {
	ID      int64
	Counter int64
}

func panicOnError(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

func naiveMethod(db *gorm.DB, num int, done chan bool) {
	for i := 0; i < num; i++ {
		var tt TestTable
		panicOnError(db.First(&tt, 1).Error)
		tt.Counter -= 1
		panicOnError(db.Save(&tt).Error)
	}
	done <- true
}

func transactionMethod(db *gorm.DB, num int, done chan bool) {
	for i := 0; i < num; i++ {
		var tt TestTable
		tx := db.Begin()
		panicOnError(tx.First(&tt, 1).Error)
		tt.Counter -= 1
		panicOnError(tx.Save(&tt).Error)
		panicOnError(tx.Commit().Error)
	}
	done <- true
}

func lockedMethod(db *gorm.DB, num int, done chan bool) {
	for i := 0; i < num; i++ {
		var tt TestTable
		tx := db.Begin()
		panicOnError(tx.Raw("SELECT * FROM test_table WHERE id = 1 LIMIT 1 FOR UPDATE").Scan(&tt).Error)
		tt.Counter -= 1
		panicOnError(tx.Save(&tt).Error)
		panicOnError(tx.Commit().Error)
	}
	done <- true
}

func atomicMethod(db *gorm.DB, num int, done chan bool) {
	for i := 0; i < num; i++ {
		panicOnError(db.Exec("UPDATE test_table SET counter = counter - 1").Error)
	}
	done <- true
}

var methodMap = map[string]func(*gorm.DB, int, chan bool){
	"naive":       naiveMethod,
	"transaction": transactionMethod,
	"locked":      lockedMethod,
	"atomic":      atomicMethod,
}

var methodDoc = map[string]string{
	"naive": `
SELECT * FROM table;
UPDATE table SET ...;
	`,
	"transaction": `
START TRANSACTION;
SELECT * FROM table;
UPDATE table SET ...;
COMMIT;
	`,
	"locked": `
START TRANSACTION;
SELECT * FROM table FOR UPDATE;
UPDATE table SET ...;
COMMIT;
	`,
	"atomic": `
UPDATE table SET counter = counter - 1;
	`,
}

func main() {
	flag.Parse()
	m := methodMap[*method]

	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True")
	if err != nil {
		logrus.Fatal("Database open error: ", err)
	}
	if err := db.DB().Ping(); err != nil {
		logrus.Fatal("Database ping error: ", err)
	}
	db.SingularTable(true)
	db.DropTable(&TestTable{})
	panicOnError(db.AutoMigrate(&TestTable{}).Error)

	c := int64(*iterNum * *goNum)
	panicOnError(db.Save(&TestTable{Counter: c}).Error)

	logrus.Infof("Initial counter state: %d", c)
	logrus.Infof("Number of decrements per goroutine: %d", *iterNum)
	logrus.Infof("Number of goroutines: %d", *goNum)
	logrus.Infof("Method: %s", *method)
	logrus.Infof("Method description:\n%s", strings.TrimSpace(methodDoc[*method]))
	done := make(chan bool)
	for i := 0; i < *goNum; i++ {
		go m(&db, *iterNum, done)
	}
	for i := 0; i < *goNum; i++ {
		<-done
	}

	var tt TestTable
	panicOnError(db.First(&tt, 1).Error)
	logrus.Infof("Result: %d", tt.Counter)
}
