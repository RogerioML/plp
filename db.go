package plp

import (
	"database/sql"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	db   *sql.DB
	muRd sync.Mutex
	rd   redis.Conn
	Pool *redis.Pool
)

// IniDb incia as variáveis do DB
func IniDb(
	dsn string,
	maxConn int,
) error {
	var err error
	db, err = sql.Open("godror", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	db.SetMaxOpenConns(maxConn)
	return nil
}

// IniRedis inicia as variáveis do REDIS
func IniRedis(address string) error {
	var err error
	rd, err = redis.Dial("tcp", address)
	if err != nil {
		return err
	}
	Pool = &redis.Pool{
		MaxIdle:         64,
		IdleTimeout:     240 * time.Second,
		MaxConnLifetime: 300 * time.Second,
		Wait:            true,
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial("tcp", address)
		},
	}
	return rd.Err()
}

func toDate(t time.Time) string {
	return t.Format("02012006150405")
}
