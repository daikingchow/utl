package utl

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2/bson"
	"log"

	"reflect"
	"strconv"
)

var (
	ErrUnSupportedType error = errors.New("Unsupported type error")
	ErrNotExist        error = errors.New("Not exist error")
)

// 重写生成连接池方法
func newPool(uri string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", uri)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

// redis store
type RedisStore struct {
	uri    string //redis服务器连接地址
	demain string //作用域
	pool   *redis.Pool
}

func NewRedisStore(uri string, demain string) *RedisStore {

	store := &RedisStore{uri: uri, pool: newPool(uri), demain: demain}
	return store
}

func (s *RedisStore) Put(key string, value interface{}) (err error) {
	key = s.demain + key
	var data interface{}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice:
		{
			data, err = bson.Marshal(value)
			if err != nil {
				log.Fatal(err)
				return err
			}
		}
	default:
		data = value

	}

ReTry:
	conn := s.pool.Get()
	defer conn.Close()
	if _, err = conn.Do("SET", key, data); err != nil {
		log.Println("Put ERROR:", err.Error())
		if err == nil {
			goto ReTry
		}
	}

	return nil
}

func (s *RedisStore) GetObject(key string, out interface{}) error {
	key = s.demain + key

ReTry:
	conn := s.pool.Get()
	defer conn.Close()
	ret, err := conn.Do("GET", key)
	if err != nil {
		log.Println("GetObject ERROR:", err.Error())
		if err == nil {
			goto ReTry
		}
	}

	if ret != nil {
		err = bson.Unmarshal(ret.([]byte), out)
		return nil
	}
	return ErrNotExist
}

func (s *RedisStore) GetString(key string) (string, error) {
	key = s.demain + key
ReTry:
	conn := s.pool.Get()
	defer conn.Close()
	ret, err := conn.Do("GET", key)
	if err != nil {
		log.Println("GetString ERROR:", err.Error())
		if err == nil {
			goto ReTry
		}
		return "", err
	}

	if ret == nil {
		return "", ErrNotExist
	}
	return string(ret.([]byte)), nil
}

func (s *RedisStore) GetMustString(key string) string {

	str, _ := s.GetString(key)
	return str
}

func (s *RedisStore) GetInt(key string) (int, error) {

	ret, err := s.GetString(key)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(ret)
}

func (s *RedisStore) GetMustInt(key string) int {

	ret, err := s.GetString(key)
	if err != nil {
		return 0
	}

	num, _ := strconv.Atoi(ret)
	return num
}

func (s *RedisStore) GetInt64(key string) (int64, error) {

	ret, err := s.GetString(key)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(ret, 10, 64)
}

func (s *RedisStore) GetMustInt64(key string) int64 {

	ret, err := s.GetString(key)
	if err != nil {
		return 0
	}

	num, _ := strconv.ParseInt(ret, 10, 64)
	return num
}

func (s *RedisStore) GetFloat32(key string) (float32, error) {

	ret, err := s.GetString(key)
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseFloat(ret, 32)
	return float32(num), err
}

func (s *RedisStore) GetMustFloat32(key string) float32 {

	ret, err := s.GetString(key)
	if err != nil {
		return 0
	}

	num, _ := strconv.ParseFloat(ret, 64)
	return float32(num)
}

func (s *RedisStore) GetFloat64(key string) (float64, error) {

	ret, err := s.GetString(key)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(ret, 64)
}

func (s *RedisStore) GetMustFloat64(key string) float64 {

	ret, err := s.GetString(key)
	if err != nil {
		return 0
	}

	num, _ := strconv.ParseFloat(ret, 64)
	return num
}

func (s *RedisStore) Del(key string) error {
	key = s.demain + key

	conn := s.pool.Get()
	defer conn.Close()
	// ret为受影响的记录数
	ret, e := conn.Do("DEL", key)
	log.Println(ret)
	return e
}
