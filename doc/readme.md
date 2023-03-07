# REDIS

## SET
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


