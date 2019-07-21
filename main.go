package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/kshvakov/clickhouse"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	connect, err := sql.Open("clickhouse", "tcp://localhost:9000?debug=true")
	if err != nil {
		log.Fatal(err)
	}
	if err := connect.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return
	}

	insertMetric(connect)

}

func insertMetric(connect *sql.DB) {
	var tx, _ = connect.Begin()
	stmt, err := tx.Prepare("INSERT INTO dev.metric values (?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	var allArgs [][]interface{}
	var minute = time.Now().Unix() / 60
	var metricCount = 1000000
	var countEveryMetric = 10
	for i := 0; i < metricCount; i++ {
		var metric = uuid.NewV4().String()
		for j := 0; j < countEveryMetric; j++ {
			allArgs = append(allArgs, []interface{}{minute, metric, rand.Intn(10)})
		}

	}

	var n = time.Now().Unix()
	for i := 0; i < len(allArgs); i++ {
		if _, err := stmt.Exec(allArgs[i]...); err != nil {
			log.Fatal(err)
		}
	}
	var prepare = time.Now().Unix() - n

	n = time.Now().Unix()
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	var commit = time.Now().Unix() - n
	fmt.Printf("1C/8G docker，插入3列、%d行，prepare %d秒, commit %d秒\n",
		metricCount*countEveryMetric, prepare, commit)
}

func insertOnTime(connect *sql.DB) {
	var err error
	var scanner *bufio.Scanner
	for _, y := range []int{2017, 2018} {
		for i := 1; i <= 12; i++ {
			var file = fmt.Sprintf("onTime_%d_%d.csv", y, i)

			if f, err := os.Open(file); err != nil {
				panic(err)
			} else {
				scanner = bufio.NewScanner(f)
			}

			var size int
			scanner.Scan()
			var head = trimComma(scanner.Text())
			var marks []string
			for range strings.Split(head, ",") {
				marks = append(marks, "?")
			}

			var tx, _ = connect.Begin()
			var stmt *sql.Stmt

			var sql = fmt.Sprintf("INSERT INTO dev.ontime values (%s)", strings.Join(marks, ","))
			if stmt, err = tx.Prepare(sql); err != nil {
				log.Fatal(err)
			}

			var total int
			var allArgs [][]interface{}
			for scanner.Scan() {
				var args []interface{}
				var line = trimComma(scanner.Text())
				for i, arg := range strings.Split(line, ",") {
					args = append(args, trans(types[i], arg))
				}

				allArgs = append(allArgs, args)
				total++
				size += len(line)
			}

			var n = time.Now().Unix()
			for i := 0; i < len(allArgs); i++ {
				if _, err := stmt.Exec(allArgs[i]...); err != nil {
					log.Fatal(err)
				}
			}
			var prepare = time.Now().Unix() - n

			n = time.Now().Unix()
			if err := tx.Commit(); err != nil {
				log.Fatal(err)
			}

			var commit = time.Now().Unix() - n
			fmt.Printf("1C/8G docker，插入100列、%d行，文件 %s, 共%dM，平均每行%dbytes，prepare %d秒, commit %d秒\n",
				total, file, size/(1024*1024), size/total, prepare, commit)
		}
	}
}

func trimComma(s string) string {
	s = strings.Replace(strings.TrimSuffix(s, ","), ".00", "", -1)
	s = strings.Replace(s, ", ", "", -1)
	return s
}

func randomMetricId() string {
	return fmt.Sprintf("mymetric-%d", rand.Intn(10))
}

func randomN() int {
	return rand.Intn(100)
}

func currentMinute() int64 {
	return time.Now().Unix() / 60
}

func currentDate() int64 {
	return time.Now().Unix() / (24 * 3600)
}

var types = []string{
	"UInt16",
	"UInt8",
	"UInt8",
	"UInt8",
	"UInt8",
	"Date",
	"FixedString(7)",
	"Int32",
	"FixedString(2)",
	"String",
	"String",
	"Int32",
	"Int32",
	"Int32",
	"FixedString(5)",
	"String",
	"FixedString(2)",
	"String",
	"String",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"FixedString(5)",
	"String",
	"FixedString(2)",
	"String",
	"String",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"String",
	"String",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"String",
	"UInt8",
	"FixedString(1)",
	"UInt8",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"UInt8",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"Int32",
	"String",
	"String",
	"String",
	"String",
	"String",
	"String",
	"String",
	"String",
	"String",
	"Int32",
	"Int32",
	"String",
	"String",
	"String",
	"String",
	"String",
	"String",
	"Int32",
	"Int32",
	"String",
	"String",
	"String",
	"String",
	"String",
	"String",
	"Int32",
	"Int32",
	"String",
	"String",
	"String",
	"String",
	"String",
	"String",
	"Int32",
	"Int32",
	"String",
	"String",
	"String",
	"String",
	"String",
	"String",
	"Int32",
	"Int32",
	"String",
	"String",
	"String",
	"String",
	"String",
}

func trans(typ string, v string) interface{} {
	var vv = strings.ReplaceAll(v, `"`, "")
	if strings.Contains(typ, "nt") {
		if vv == "" || vv == "0" {
			return 0
		}

		if n, err := strconv.Atoi(strings.TrimPrefix(vv, "0")); err != nil {
			panic(fmt.Sprintf("parse %s %s err: %s", v, vv, err))
		} else {
			return n
		}
	}

	return vv

}
