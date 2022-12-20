//go:build mongodb

package rate

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"

	pb "github.com/ucy-coast/hotel-app/internal/rate/proto"
	"github.com/ucy-coast/hotel-app/pkg/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"time"

	"github.com/bradfitz/gomemcache/memcache"

)

type DatabaseSession struct {

	MongoSession *mgo.Session
	MemcClient   *memcache.Client
}

func NewDatabaseSession(db_addr string, memc_addr string) *DatabaseSession {
	// TODO: Implement me

	session, err := mgo.Dial(db_addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("New session successfull...")

	memc_client := memcache.New(memc_addr)
	memc_client.Timeout = time.Second * 2
	memc_client.MaxIdleConns = 512

	return &DatabaseSession{
		MongoSession: session,
		MemcClient: memc_client,
	}
}

func (db *DatabaseSession) LoadDataFromJsonFile(rateJsonPath string) {
	util.LoadDataFromJsonFile(db.MongoSession, "rate-db", "inventory", rateJsonPath)
}

// GetRates gets rates for hotels for specific date range.
func (db *DatabaseSession) GetRates(hotelIds []string) (RatePlans, error) {
	// TODO: Implement me


	rates := make([]*pb.RatePlan, 0)

	for _, id := range hotelIds {
		// first check memcached

		item, err := db.MemcClient.Get(id)
		if err == nil {
			// memcached hit
			log.Infof("Memcached hit: hotel_id == %v\n", id)
			hotel_rate := new(pb.RatePlan)
			if err = json.Unmarshal(item.Value, hotel_rate); err != nil {
				log.Warn(err)
			}
			rates = append(rates, hotel_rate)
		} else if err == memcache.ErrCacheMiss {
			// memcached miss, set up mongo connection
			log.Infof("Memcached miss: hotel_id == %v\n", id)
			session := db.MongoSession.Copy()
			defer session.Close()
			c := session.DB("rate-db").C("inventory")
		
			hotel_rate := new(pb.RatePlan)
			err := c.Find(bson.M{"id": id}).One(&hotel_rate)
			if err != nil {
				log.Fatalf("Failed get hotels rate plan: ", err)
			}
			rates = append(rates, hotel_rate)

			rate_json, err := json.Marshal(hotel_rate)
			if err != nil {
				log.Errorf("Failed to marshal hotel rate [id: %v] with err:", hotel_rate.HotelId, err)
			}
			memc_str := string(rate_json)

			// write to memcached
			err = db.MemcClient.Set(&memcache.Item{Key: id, Value: []byte(memc_str)})
			if err != nil {
				log.Warn("MMC error: ", err)
			}
		} else {
			log.Errorf("Memcached error = %s\n", err)
			panic(err)
		}	
	}





	return rates, nil


}
