# REDIS

## INIT
```go
package main

import (
	"encoding/json"
	"fmt"
)

// import  "strconv"
// import "sync"
import "time"
import "github.com/go-redis/redis"

// Сессия подключения
var client *redis.Client

// **************************************************
// Подключение к базе данных Redis
// **************************************************
func init() {

	client = redis.NewClient(&redis.Options{
		Addr:         ":6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		DB:           2,
	})

	// Очистка
	//client.FlushDB()
}

type Mst map[string]interface{}

type Corp struct {
	Title string `json:"title"`
	Name  string `json:"name"`
	Form  string `json:"form"`
}

func main() {
	Iterrator()
}
```


## MSET
```go
unc SettMain() {
	fmt.Println("CLIENT ID", client.ClientID())
	fmt.Println("CLIENT ID", client.DBSize())

	ol := `{
         Id:"Key-001",
         Name:"Roman",
         Description:"Samples records"
    }
    `
	// Запись одной командой нескольких значений
	client.MSet("one", "1-3663", "two", "2", "three", "3", "four", "4", "Pole:12", "12", "Pole:14", "14", "Pole:15", "15")

	// Запись одного значения
	err := client.Set("Keys:15", "key15 значение", 0).Err()
	if err != nil {
		panic(err)
	}

	// Чтение одного значения
	val, _ := client.Get("Keys:15").Result()
	fmt.Println("15 елемент ", val)

	// Запись одного значения
	client.Set("Structure", ol, 0)

	// выбор по фильтру
	keys := client.Keys("*o*")
	fmt.Println("Выбор по фильтру", keys)

	// выбор по фильтру
	// on - начало поиска
	// ? - количество символов после начала
	keys = client.Keys("on?")
	fmt.Println("выбор по фильтру ", keys)

	// Все записи
	keys = client.Keys("*")
	fmt.Println("Все записи ", keys)

	// Все записи
	keys = client.Keys("Pole:*")
	fmt.Println("Все поля Pole", keys)

	// Установка время жизни для опредленного ключа
	client.Expire("Pole:14", 25*time.Second)

	ttl := client.TTL("Pole:14")
	fmt.Println("Опредляем время жизни переменной ", ttl)

	// Имеет ли ключ постоянное значение или временное
	persist := client.Persist("Pole:12")
	fmt.Println("Persist : ", persist)

	expireAt := client.ExpireAt("key", time.Now().Add(-time.Hour))
	fmt.Println("Response : ", expireAt.Val())

	// перенос ключа в базу 1
	move := client.Move("Pole:15", 1)
	fmt.Println("Move :", move)

	// Установка со временем в секундах
	expiration := 90 * time.Second
	client.Set("Status", "Еуые", expiration)

	// Удаление
	// Deleting()
}
```


## Set for

```go

// *************************************
// Загрузка записей
// *************************************
func AllOpertaion() {

	// Добавление значения
	client.Set("key2", "First element", 0)
	client.Set("key3", "New element", 0)
	client.Set("key2:456", "Example simple for insert", 0)

	err := client.Set("key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get("key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()

	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
}
```


## HSET
![image](https://user-images.githubusercontent.com/3950155/223431990-226e6c8c-ee84-4c81-876f-7cf14d28a57e.png)

```go
// Main
func HMSET() {

	//fmt.Println(client.ClientList())

	// Запись списком
	dat := Mst{}
	dat["a1"] = "Фамилия"
	dat["a3"] = "Имя"
	dat["a4"] = "Other"
	dat["passp"] = "Паспорт"

	client.HMSet("mst:001", dat)

	// Чтение только перечисленных полей
	// Если не указать поля то их не будет в результатах вывода
	cle := client.HMGet("mst:001", "a1", "a3", "a4", "passp")

	fmt.Println("Имя команды ", cle.Name())
	fmt.Println("Значения в массиве ", cle.Val())
	fmt.Println("Аргументы в массиве ", cle.Args())
	fmt.Println("Аргументы в символьный ", cle.String())

	res, _ := cle.Result()
	fmt.Println("Результат в массиве ", res)

	// Loop results
	for _, rr := range res {
		fmt.Println("Result :", rr)
	}
}
```


## HGET
```go
func HSETHGET() {

	// Добавление в список по одному
	client.HSet("mst:001:01:ART", "system", 1)
	client.HSet("mst:001:01:ART", "other", "Прочие фитчи")
	client.HSet("mst:001:01:ART", "prim", "Примечание")
	client.HSet("mst:001:01:ART", "settings", "Установки примечание")

	ss := client.HGet("mst:001:01:ART", "system")
	fmt.Println(ss)

	ss = client.HGet("mst:001:01:ART", "other")
	fmt.Println(ss)

	ss = client.HGet("mst:001:01:ART", "prim")
	fmt.Println(ss)

	// Удаление из смпика по ключу
	client.HDel("mst:001:01:ART", "prim")

	// Проверка наличие в списке по ключу
	ye := client.HExists("mst:001:01:ART", "prim")
	fmt.Println("Has prim ", ye)

	// Проверка наличие в списке по ключу
	ye = client.HExists("mst:001:01:ART", "other")
	fmt.Println("Has other : ", ye)

	keys := client.HGetAll("mst:001:01:ART")
	fmt.Println("Vals    : ", keys.Val())
	fmt.Println("Key     : ", keys.Val()["other"])
	fmt.Println("Args    : ", keys.Args())
	fmt.Println("Strings : ", keys.String())

	res, _ := keys.Result()
	fmt.Println("Results : ", res)
	fmt.Println("Result : ", res["other"])

	// Получение всех значений в списке
	for i, rr := range keys.Val() {
		fmt.Printf("ELEM %s Name : %s \n", i, rr)
	}

	// Получение всех значений в списке
	for i, rr := range keys.Args() {
		fmt.Println("АРГ :", i, rr)
	}
}
```


## JSON
```go

// Set Get json
func Set_Json() {

	// Хранение 10 сек
	dur := time.Duration(time.Second * 10)

	c := Corp{
		Title: "Title",
		Name:  "Name company",
		Form:  "Форма",
	}

	ff, _ := json.Marshal(c)

	// Запись одного значения
	err := client.Set("jss", ff, dur)
	if err != nil {
		fmt.Println(err)
	}

	// Read Json
	kl, _ := client.Get("jss").Bytes()
	json.Unmarshal(kl, &c)

	// Old variant
	// val, _ := client.Get("jss").Result()
	// json.Unmarshal([]byte(val), &c)

	fmt.Println(c)
	fmt.Println(c.Title)
	fmt.Println(c.Name)

}
```




## DELETE
```go
// *************************************
// Удаление ключей
// *************************************
func Deleting() {
	n, _ := client.Del("key1", "key2", "key3").Result()
	fmt.Println(n)
}
```



## PUBLISH - SUBSCRIBE
```go
// Pub Sub in chanel
func Pub_Sub() {

	fmt.Println("Start ... ")
	go Pub("mchanel")
	go Pub("mc1")

	go Subb("mchanel")
	Subb("mc1")
}

// Publish
func Pub(ch string) {
	fmt.Println("PUBLISH :", ch)
	c := 0

	for i := 0; i < 10; i++ {
		c = i
		err := client.Publish(ch, fmt.Sprintf("payload ID  %v", c)).Err()
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Millisecond * 300)
	}
	time.Sleep(time.Millisecond * 1000)

	c++
	client.Publish(ch, fmt.Sprintf("Окончание ID  %v", c)).Err()
	time.Sleep(time.Millisecond * 300)

	c++
	client.Publish(ch, fmt.Sprintf("Передача события 2 ID  %v", c)).Err()
	time.Sleep(time.Millisecond * 300)
}

// Subscribe
func Subb(ch string) {
	fmt.Println("SUBSCRIBE :", ch)

	pubsub := client.Subscribe(ch)
	//defer pubsub.Close()
	//pubsub.Receive()

	c := pubsub.Channel()
	for msg := range c {
		fmt.Println(msg.Channel, msg.Payload)
	}
}
```

## FIND IN FOR
```go

// Iterator
func Iterrator() {
	// Set_Json()
	// Set_value_loop()

	// https://redis.uptrace.dev/guide/get-all-keys.html#cluster-and-ring
	// kk:1, kk:2
	iter := client.Scan(0, "kk:*", 0).Iterator()

	for iter.Next() {
		fmt.Println("keys", iter.Val())
	}

	if err := iter.Err(); err != nil {
		panic(err)
	}

}
```



