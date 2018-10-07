package conn

import (
	"github.com/devfeel/database/mysql"
)

var mysqlClient *mysql.MySqlDBContext

type MysqlConn struct{}

func init() {
	mysqlClient = mysql.NewMySqlDBContext("root:123456@tcp(118.31.32.168:3306)/polaris?charset=utf8&allowOldPasswords=1")

}

//get mysqlClient conn
func GetMysqlClient() *mysql.MySqlDBContext {
	if mysqlClient == nil {
		panic("redis connection failed!")
	}
	return mysqlClient
}
