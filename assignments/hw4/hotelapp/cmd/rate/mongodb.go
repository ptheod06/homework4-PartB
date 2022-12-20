//go:build mongodb

package main

import (
	"flag"
	"path/filepath"

	"github.com/ucy-coast/hotel-app/internal/rate"
)

var (
	database_addr = flag.String("db_addr", "mongodb-rate:27017", "Address of the data base server")
	memcache_addr = flag.String("memc_addr", "memcached-profile:11211", "Address of the memcache server")
)

func initializeRateDatabase() *rate.DatabaseSession {
	dbsession := rate.NewDatabaseSession(*database_addr, *memcache_addr)
	dbsession.LoadDataFromJsonFile(filepath.Join(*jsonDataDir, "inventory.json"))
	return dbsession
}
