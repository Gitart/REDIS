
package main
import "fmt"
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
        DB:2,
    })
    client.FlushDB()
}




// ***********************************************************
// Процедура
// ***********************************************************
func main(){

    ol:=`{
         Id:"Key-001",
         Name:"Roman",
         Description:"Samples records"
    }
    `
    // Запись в одной строчке нескольких значений
    client.MSet("one", "1-3663", "two", "2", "three", "3", "four", "4", "Pole:12", "12", "Pole:14", "14","Pole:15", "15")

    // Запись в одного значения
    client.Set("Structure",     ol, 0)

    // выбор по фильтру
    keys := client.Keys("*o*")
    fmt.Println(keys)

    // выбор по фильтру 
    // on - начало поиска
    // ? - количество символов после начала
    keys = client.Keys("on?")
    fmt.Println(keys)
    
    // Все записи
    keys = client.Keys("*")
    fmt.Println(keys)

    // Все записи
    keys = client.Keys("Pole:*")
    fmt.Println("Все поля ",keys)

    // Установка время жизни для опредленного ключа
    client.Expire("Pole:14", 25*time.Second)
    
    ttl := client.TTL("Pole:14")
    fmt.Println("Опредляем время жизни переменной ",ttl)

    // Имеет ли ключ постоянное значение или временное
    persist := client.Persist("Pole:12")
    fmt.Println("Persist : ", persist) 


    expireAt := client.ExpireAt("key", time.Now().Add(-time.Hour))
    fmt.Println("Response : ",expireAt.Val())

    // перенос ключа в базу 1
    move := client.Move("Pole:15", 1)
    fmt.Println("Move :", move)     


    // Установка со временем в секундах
    expiration := 90 * time.Second
    client.Set("Status", "Еуые", expiration)

    // Удаление
    Deleting()
}





// *************************************
// 
// *************************************
func Deleting() {
    n, _ := client.Del("key1", "key2", "key3").Result()
    fmt.Println(n)
}


// *************************************
// 
// *************************************
func Testing(){

     // Загрузка данных
     LoadBillonRecords(100)

     // Чтение 15 елемента
     val, err := client.Get("Keys:15").Result()
     if err != nil {         
        panic(err.Error())
     }
    fmt.Println(val)

   v:=client.Set("key32","Другой", 1000000)
   fmt.Println(v.String())

}

// *************************************
// Загрузка записей
// *************************************
func AllOpertaion(){
      // Добавление значения
       client.Set("key2",     "Другой", 0)
       client.Set("key3",     "Новости пример", 0)
       client.Set("key2:456", "Вот пример простой", 0)

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



// *************************************
// Загрузка миллиона записей
// *************************************
func LoadBillonRecords(Recs int){
     fstart:=time.Now()

     for i:=1; i<Recs; i++{
         f:="Keys:"+InttoStr(i)
         client.Set(f, "value2222  ssss", 0)    
     }


     // Замер времени
     finsh:=time.Now()
     ffinsh:=fstart.Sub(finsh)
     fmt.Println("Operation duration : ",ffinsh)
 }


/******************************************************************
 * Конвертация Int to Str
 ******************************************************************/
func InttoStr(Ints int) string {
    //str := strconv.FormatInt(Intt64, 10)      // Выдает конвертацию 2000-wqut
    //str := strconv.Itoa64(Int64)              // use base 10 for sanity purpose
    str := fmt.Sprintf("%d", Ints)
    return str
}








