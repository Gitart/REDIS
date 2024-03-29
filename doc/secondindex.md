# Вторичная индексация

Создание вторичных индексов в Redis

**Redis** — это не совсем хранилище ключей и значений, поскольку значения могут быть сложными структурами данных. Однако у него есть внешняя оболочка «ключ-значение»: на уровне API данные адресуются по имени ключа. Справедливо сказать, что изначально Redis предлагает _доступ только к первичному ключу_ . Однако, поскольку Redis является сервером структур данных, его возможности можно использовать для индексации, чтобы создавать вторичные индексы разных типов, в том числе составные (многостолбцовые) индексы.

В этом документе объясняется, как можно создавать индексы в Redis, используя следующие структуры данных:

* Отсортированные наборы для создания вторичных индексов по идентификатору или другим числовым полям.   
* Отсортированные наборы с лексикографическими диапазонами для создания более сложных вторичных индексов, составных индексов и индексов обхода графа.   
* Наборы для создания случайных индексов.   
* Списки для создания простых итерируемых индексов и индексов последних N элементов.        

Внедрение и поддержка индексов с помощью **Redis** — сложная тема, поэтому большинство пользователей, которым необходимо выполнять сложные запросы к данным, должны понимать, лучше ли их обслуживает реляционное хранилище. Однако часто, особенно в сценариях кэширования, существует явная необходимость хранить проиндексированные данные в Redis, чтобы ускорить общие запросы, для выполнения которых требуется какая-либо форма индексирования.

# Простые числовые индексы с отсортированными наборами

Самый простой вторичный индекс, который вы можете создать с помощью Redis, — это использовать тип данных отсортированного набора, который представляет собой структуру данных, представляющую набор элементов, упорядоченных по числу с плавающей запятой, которое является оценкой _каждого_ элемента. Элементы упорядочены от наименьшего к наибольшему количеству баллов.

Поскольку оценка представляет собой число с двойной точностью, индексы, которые вы можете построить с помощью ванильных отсортированных наборов, ограничены вещами, в которых поле индексации представляет собой число в заданном диапазоне.

Две команды для создания таких индексов — это [`ZADD`](/commands/zadd)и [`ZRANGE`](/commands/zrange)с `BYSCORE`аргументом для добавления элементов и извлечения элементов в указанном диапазоне соответственно.

Например, можно проиндексировать набор имен людей по их возрасту, добавив элемент в отсортированный набор. Элементом будет имя человека, а баллом будет возраст.

```
ZADD myindex 25 Manuel
ZADD myindex 18 Anna
ZADD myindex 35 Jon
ZADD myindex 67 Helen
```

Чтобы получить всех лиц в возрасте от 20 до 40 лет, можно использовать следующую команду:

```
ZRANGE myindex 20 40 BYSCORE
1) "Manuel"
2) "Jon"
```

С помощью параметра **WITHSCORES**[`ZRANGE`](/commands/zrange) также можно получить оценки, связанные с возвращенными элементами.

Команду [`ZCOUNT`](/commands/zcount)можно использовать для получения количества элементов в заданном диапазоне без фактической выборки элементов, что также полезно, особенно с учетом того факта, что операция выполняется за логарифмическое время независимо от размера диапазона.

Диапазоны могут быть как включающими, так и исключающими. [`ZRANGE`](/commands/zrange) Дополнительные сведения см. в документации по командам.

**Примечание** . Используя [`ZRANGE`](/commands/zrange)с аргументами `BYSCORE`и `REV`, можно запрашивать диапазон в обратном порядке, что часто полезно, когда данные индексируются в заданном направлении (по возрастанию или по убыванию), но мы хотим получить информацию в обратном порядке.

## Использование идентификаторов объектов в качестве связанных значений

В приведенном выше примере мы связали имена с возрастом. Однако в общем случае мы можем захотеть проиндексировать какое-то поле объекта, которое хранится в другом месте. Вместо того, чтобы использовать значение отсортированного набора напрямую для хранения данных, связанных с индексированным полем, можно сохранить только идентификатор объекта.

Например, у меня могут быть хэши Redis, представляющие пользователей. Каждый пользователь представлен одним ключом, напрямую доступным по ID:

```
HMSET user:1 id 1 username antirez ctime 1444809424 age 38
HMSET user:2 id 2 username maria ctime 1444808132 age 42
HMSET user:3 id 3 username jballard ctime 1443246218 age 33
```

Если я хочу создать индекс, чтобы запрашивать пользователей по их возрасту, я мог бы сделать:

```
ZADD user.age.index 38 1
ZADD user.age.index 42 2
ZADD user.age.index 33 3
```

На этот раз значение, связанное с оценкой в ​​отсортированном наборе, является идентификатором объекта. Поэтому, когда я запрашиваю индекс с [`ZRANGE`](/commands/zrange)помощью аргумента `BYSCORE`, мне также нужно будет получить необходимую информацию с помощью [`HGETALL`](/commands/hgetall)или подобных команд. Очевидным преимуществом является то, что объекты могут изменяться, не касаясь индекса, пока мы не меняем индексированное поле.

В следующих примерах мы почти всегда будем использовать идентификаторы в качестве значений, связанных с индексом, так как обычно это более здравый подход, за некоторыми исключениями.

## Обновление индексов простых отсортированных наборов

Часто мы индексируем вещи, которые со временем меняются. В приведенном выше примере возраст пользователя меняется каждый год. В таком случае имело бы смысл использовать в качестве индекса дату рождения вместо самого возраста, но есть и другие случаи, когда мы просто хотим, чтобы какое-то поле время от времени менялось, а индекс отражал это изменение.

Эта [`ZADD`](/commands/zadd)команда делает обновление простых индексов очень тривиальной операцией, поскольку повторное добавление элемента с другой оценкой и тем же значением просто обновит оценку и переместит элемент в нужное положение, поэтому, если пользователю исполнилось 39 лет, `antirez`чтобы чтобы обновить данные в хэше, представляющем пользователя, а также в индексе, нам нужно выполнить следующие две команды:

```
HSET user:1 age 39
ZADD user.age.index 39 1
```

Операция может быть заключена в транзакцию [`MULTI`](/commands/multi)/ [`EXEC`](/commands/exec), чтобы убедиться, что оба поля обновлены или нет.

## Преобразование многомерных данных в линейные данные

Индексы, созданные с помощью отсортированных наборов, могут индексировать только одно числовое значение. Из-за этого вы можете подумать, что невозможно проиндексировать что-то, что имеет несколько измерений, используя такие индексы, но на самом деле это не всегда так. Если вы можете эффективно представить что-то многомерное линейным способом, часто можно использовать простой отсортированный набор для индексации.

Например, [API геоиндексации Redis](/commands/geoadd) использует отсортированный набор для индексации мест по широте и долготе с использованием метода, называемого [Geo hash](https://en.wikipedia.org/wiki/Geohash) . Оценка отсортированного набора представляет чередующиеся биты долготы и широты, так что мы сопоставляем линейную оценку отсортированного набора со многими маленькими _квадратами_ на поверхности земли. Выполняя поиск в стиле 8+1 по центру и окрестностям, можно получить элементы по радиусу.

## Пределы счета

Счета отсортированных элементов набора являются числами с двойной точностью. Это означает, что они могут представлять разные десятичные или целые значения с разными ошибками, поскольку внутри они используют экспоненциальное представление. Однако для целей индексации интересно то, что оценка всегда может представлять без каких-либо ошибок числа между -9007199254740992 и 9007199254740992, то есть `-/+ 2^53`.

При представлении гораздо больших чисел вам нужна другая форма индексации, которая может индексировать числа с любой точностью, называемая лексикографическим индексом.

# Лексикографические указатели

Отсортированные множества Redis обладают интересным свойством. Когда добавляются элементы с одинаковой оценкой, они сортируются лексикографически, сравнивая строки как двоичные данные с функцией `memcmp()`.

Для людей, которые не знают ни языка C, ни `memcmp`функции, это означает, что элементы с одинаковым счетом сортируются, сравнивая необработанные значения их байтов, байт за байтом. Если первый байт совпадает, проверяется второй и так далее. Если общий префикс двух строк одинаков, то более длинная строка считается большей из двух, поэтому «foobar» больше, чем «foo».

Существуют такие команды, как [`ZRANGE`](/commands/zrange)и [`ZLEXCOUNT`](/commands/zlexcount), которые могут запрашивать и подсчитывать диапазоны лексикографически, при условии, что они используются с отсортированными наборами, где все элементы имеют одинаковую оценку.

Эта функция Redis в основном эквивалентна структуре `b-tree`данных, которая часто используется для реализации индексов с традиционными базами данных. Как вы можете догадаться, из-за этого можно использовать эту структуру данных Redis для реализации довольно причудливых индексов.

Прежде чем мы углубимся в использование лексикографических индексов, давайте проверим, как ведут себя отсортированные множества в этом специальном режиме работы. Так как нам нужно добавлять элементы с одинаковым счетом, мы всегда будем использовать специальный счет, равный нулю.

```
ZADD myindex 0 baaa
ZADD myindex 0 abbb
ZADD myindex 0 aaaa
ZADD myindex 0 bbbb
```

Извлечение всех элементов из отсортированного набора сразу показывает, что они упорядочены лексикографически.

```
ZRANGE myindex 0 -1
1) "aaaa"
2) "abbb"
3) "baaa"
4) "bbbb"
```

Теперь мы можем использовать [`ZRANGE`](/commands/zrange)с `BYLEX`аргументом для выполнения запросов диапазона.

```
ZRANGE myindex [a (b BYLEX
1) "aaaa"
2) "abbb"
```

Обратите внимание, что в запросах диапазона мы ставили префикс `min`и `max`элементы, идентифицирующие диапазон, со специальными символами `[`и `(`. Эти префиксы являются обязательными и указывают, являются ли элементы диапазона включающими или исключающими. Таким образом, диапазон `[a (b`означает дать мне лексикографически все элементы между `a`инклюзивным и `b`исключающим, то есть все элементы, начинающиеся с `a`.

Есть также еще два специальных символа, обозначающих бесконечно отрицательную строку и бесконечно положительную строку: `-`и `+`.

```
ZRANGE myindex [b + BYLEX
1) "baaa"
2) "bbbb"
```

Вот и все. Давайте посмотрим, как использовать эти функции для построения индексов.

## Первый пример: завершение

Интересным применением индексации является завершение. Завершение — это то, что происходит, когда вы начинаете вводить свой запрос в поисковую систему: пользовательский интерфейс предвидит то, что вы, вероятно, вводите, предоставляя общие запросы, которые начинаются с одних и тех же символов.

Наивный подход к завершению состоит в том, чтобы просто добавлять каждый запрос, который мы получаем от пользователя, в индекс. Например, если пользователь выполняет поиск, `banana` мы просто сделаем:

```
ZADD myindex 0 banana
```

И так далее для каждого когда-либо встречавшегося поискового запроса. Затем, когда мы хотим завершить пользовательский ввод, мы выполняем запрос диапазона, используя [`ZRANGE`](/commands/zrange)аргумент `BYLEX`. Представьте, что пользователь вводит «бит» в форму поиска, и мы хотим предложить возможные ключевые слова для поиска, начинающиеся с «бит». Отправляем Redis такую ​​команду:

```
ZRANGE myindex "[bit" "[bit\xff" BYLEX
```

По сути, мы создаем диапазон, используя строку, которую пользователь вводит прямо сейчас, в качестве начала, и ту же строку плюс завершающий байт, установленный на 255, который `\xff`в примере является концом диапазона. Таким образом мы получаем все строки, которые начинаются со строки, которую вводит пользователь.

Обратите внимание, что мы не хотим, чтобы возвращалось слишком много элементов, поэтому мы можем использовать параметр **LIMIT** , чтобы уменьшить количество результатов.

## Добавление частоты в микс

Приведенный выше подход немного наивен, потому что таким образом все поисковые запросы пользователей одинаковы. В реальной системе мы хотим заполнять строки в соответствии с их частотой: очень популярные поисковые запросы будут предлагаться с более высокой вероятностью по сравнению с поисковыми строками, которые набираются очень редко.

Чтобы реализовать что-то, что зависит от частоты и в то же время автоматически адаптируется к будущим входам, очищая поиски, которые больше не популярны, мы можем использовать очень простой алгоритм потоковой _передачи_ .

Для начала мы модифицируем наш индекс, чтобы хранить не только поисковый запрос, но и частоту, с которой он связан. Таким образом, вместо того, чтобы просто складывать, `banana`мы добавляем `banana:1`, где 1 — частота.

```
ZADD myindex 0 banana:1
```

Нам также нужна логика для увеличения индекса, если искомый термин уже существует в индексе, поэтому на самом деле мы сделаем что-то вроде этого:

```
ZRANGE myindex "[banana:" + BYLEX LIMIT 0 1
1) "banana:1"
```

Это вернет единственную запись, `banana`если она существует. Затем мы можем увеличить соответствующую частоту и отправить следующие две команды:

```
ZREM myindex 0 banana:1
ZADD myindex 0 banana:2
```

Обратите внимание, что, поскольку возможно наличие одновременных обновлений, указанные выше три команды следует отправлять через [сценарий Lua](/commands/eval) , чтобы сценарий Lua атомарно получил старый счетчик и повторно добавил элемент с увеличенным счетом.

Таким образом, результатом будет то, что каждый раз, когда пользователь ищет, `banana`мы будем обновлять нашу запись.

Более того: наша цель состоит в том, чтобы элементы искали очень часто. Так что нам нужна какая-то форма очистки. Когда мы на самом деле запрашиваем индекс, чтобы завершить пользовательский ввод, мы можем увидеть что-то вроде этого:

```
ZRANGE myindex "[banana:" + BYLEX LIMIT 0 10
1) "banana:123"
2) "banaooo:1"
3) "banned user:49"
4) "banning:89"
```

По-видимому, никто не ищет, например, «banaooo», но запрос был выполнен один раз, поэтому мы заканчиваем представление его пользователю.

Это то, что мы можем сделать. Из возвращенных элементов мы выбираем случайный, уменьшаем его счет на единицу и повторно добавляем его с новым счетом. Однако, если оценка достигает 0, мы просто удаляем элемент из списка. Вы можете использовать гораздо более продвинутые системы, но идея состоит в том, что индекс в долгосрочной перспективе будет содержать популярные поисковые запросы, и если популярные поисковые запросы будут меняться с течением времени, он будет адаптироваться автоматически.

Усовершенствованием этого алгоритма является выбор записей в списке в соответствии с их весом: чем выше оценка, тем менее вероятные записи выбираются, чтобы уменьшить ее оценку или исключить их.

## Нормализация строк для регистра и акцентов

В примерах завершения мы всегда использовали строчные буквы. Однако на самом деле все намного сложнее: в языках имена пишутся с заглавной буквы, акценты и так далее.

Один из простых способов решить эту проблему — нормализовать строку, которую ищет пользователь. Независимо от того, что пользователь ищет «банан», «банан» или «банан», мы всегда можем превратить его в «банан».

Однако иногда мы можем захотеть предоставить пользователю исходный типизированный элемент, даже если мы нормализуем строку для индексации. Чтобы сделать это, мы изменим формат индекса, чтобы вместо простого сохранения `term:frequency`мы сохраняли, `normalized:frequency:original` как в следующем примере:

```
ZADD myindex 0 banana:273:Banana
```

По сути, мы добавляем еще одно поле, которое будем извлекать и использовать только для визуализации. Вместо этого диапазоны всегда будут вычисляться с использованием нормализованных строк. Это обычная уловка, имеющая множество применений.

## Добавление вспомогательной информации в указатель

При прямом использовании отсортированного набора у нас есть два разных атрибута для каждого объекта: оценка, которую мы используем в качестве индекса, и связанное значение. При использовании вместо этого лексикографических индексов оценка всегда устанавливается на 0 и в основном не используется вообще. У нас осталась единственная строка, которая является самим элементом.

Как и в предыдущих примерах завершения, мы по-прежнему можем хранить связанные данные с помощью разделителей. Например, мы использовали двоеточие, чтобы добавить частоту и исходное слово для завершения.

В общем, мы можем добавить любое ассоциированное значение к нашему ключу индексации. Чтобы использовать лексикографический индекс для реализации простого хранилища ключ-значение, мы просто сохраняем запись как `key:value`:

```
ZADD myindex 0 mykey:myvalue
```

И найдите ключ с помощью:

```
ZRANGE myindex [mykey: + BYLEX LIMIT 0 1
1) "mykey:myvalue"
```

Затем мы извлекаем часть после двоеточия, чтобы получить значение. Однако проблемой, которую необходимо решить в этом случае, являются коллизии. Символ двоеточия может быть частью самого ключа, поэтому его нужно выбрать, чтобы он никогда не сталкивался с добавляемым ключом.

Поскольку лексикографические диапазоны в Redis безопасны для двоичного кода, вы можете использовать любой байт или любую последовательность байтов. Однако, если вы получаете ненадежный пользовательский ввод, лучше использовать некоторую форму экранирования, чтобы гарантировать, что разделитель никогда не окажется частью ключа.

Например, если вы используете два нулевых байта в качестве разделителя `"\0\0"`, вы можете всегда экранировать нулевые байты в последовательности из двух байтов в своих строках.

## Числовое дополнение

Лексикографические индексы могут выглядеть хорошо только тогда, когда проблема состоит в том, чтобы индексировать строки. На самом деле очень просто использовать этот тип индекса для выполнения индексации чисел произвольной точности.

В наборе символов ASCII цифры появляются в порядке от 0 до 9, поэтому, если мы дополним числа слева ведущими нулями, результатом будет то, что сравнение их как строк упорядочит их по их числовому значению.

```
ZADD myindex 0 00324823481:foo
ZADD myindex 0 12838349234:bar
ZADD myindex 0 00000000111:zap

ZRANGE myindex 0 -1
1) "00000000111:zap"
2) "00324823481:foo"
3) "12838349234:bar"
```

Мы эффективно создали индекс, используя числовое поле, которое может быть сколь угодно большим. Это также работает с числами с плавающей запятой любой точности, убедившись, что мы оставили числовую часть с нулями в начале и десятичную часть с нулями в конце, как в следующем списке чисел:

```
    01000000000000.11000000000000
    01000000000000.02200000000000
    00000002121241.34893482930000
    00999999999999.00000000000000
```

## Использование чисел в двоичной форме

Хранение чисел в десятичном виде может занимать слишком много памяти. Альтернативный подход состоит в том, чтобы просто хранить числа, например 128-битные целые числа, непосредственно в их двоичной форме. Однако для того, чтобы это работало, вам нужно хранить числа в _формате с обратным порядком байтов_ , чтобы наиболее значащие байты сохранялись перед младшими байтами. Таким образом, когда Redis сравнивает строки с `memcmp()`, он эффективно сортирует числа по их значению.

Имейте в виду, что данные, хранящиеся в двоичном формате, менее заметны для отладки, их сложнее анализировать и экспортировать. Так что это определенно компромисс.

# Составные индексы

До сих пор мы исследовали способы индексации отдельных полей. Однако все мы знаем, что хранилища SQL могут создавать индексы с использованием нескольких полей. Например, я могу индексировать товары в очень большом магазине по номеру комнаты и цене.

Мне нужно выполнить запросы, чтобы получить все продукты в данной комнате с заданным ценовым диапазоном. Что я могу сделать, так это проиндексировать каждый продукт следующим образом:

```
ZADD myindex 0 0056:0028.44:90
ZADD myindex 0 0034:0011.00:832
```

Вот поля `room:price:product_id`. Я использовал только четыре цифры в этом примере для простоты. Вспомогательные данные (идентификатор продукта) не требуют заполнения.

С таким индексом получить все продукты в комнате 56 по цене от 10 до 30 долларов очень просто. Мы можем просто запустить следующую команду:

```
ZRANGE myindex [0056:0010.00 [0056:0030.00 BYLEX
```

Вышеуказанное называется составным индексом. Его эффективность зависит от порядка полей и запросов, которые я хочу выполнить. Например, приведенный выше индекс нельзя эффективно использовать для получения всех продуктов, имеющих определенный ценовой диапазон, независимо от номера комнаты. Однако я могу использовать первичный ключ для выполнения запросов независимо от цены, например, _дать мне все продукты в комнате 44_ .

Составные индексы очень эффективны и используются в традиционных хранилищах для оптимизации сложных запросов. В Redis они могут быть полезны как для реализации очень быстрого индекса Redis в памяти для чего-либо, хранящегося в традиционном хранилище данных, так и для прямого индексирования данных Redis.

# Обновление лексикографических указателей

Значение индекса в лексикографическом индексе может быть довольно причудливым и сложным или медленным для восстановления из того, что мы храним об объекте. Таким образом, один из подходов к упрощению обработки индекса за счет использования большего объема памяти состоит в том, чтобы наряду с отсортированным набором, представляющим индекс, использовать хэш, отображающий идентификатор объекта в текущее значение индекса.

Так, например, когда мы индексируем, мы также добавляем к хешу:

```
MULTI
ZADD myindex 0 0056:0028.44:90
HSET index.content 90 0056:0028.44:90
EXEC
```

Это не всегда нужно, но упрощает операции по обновлению индекса. Чтобы удалить старую информацию, которую мы проиндексировали для идентификатора объекта 90, независимо от _текущих_ значений полей объекта, нам просто нужно получить хеш-значение по идентификатору объекта и [`ZREM`](/commands/zrem)в представлении отсортированного множества.

# Представление и запрос графов с помощью hexastore

Одна интересная особенность составных индексов заключается в том, что они удобны для представления графиков с использованием структуры данных, которая называется [Hexastore](http://www.vldb.org/pvldb/vol1/1453965.pdf) .

Гексастор обеспечивает представление отношений между объектами, образованными _субъектом_ , _предикатом_ и _объектом_ . Простым отношением между объектами может быть:

```
antirez is-friend-of matteocollina
```

Чтобы представить это отношение, я могу сохранить в своем лексикографическом указателе следующий элемент:

```
ZADD myindex 0 spo:antirez:is-friend-of:matteocollina
```

Обратите внимание, что я поставил перед своим элементом строку **spo** . Это означает, что элемент представляет подлежащее, сказуемое, объектное отношение.

В можно добавить еще 5 записей для того же отношения, но в другом порядке:

```
ZADD myindex 0 sop:antirez:matteocollina:is-friend-of
ZADD myindex 0 ops:matteocollina:is-friend-of:antirez
ZADD myindex 0 osp:matteocollina:antirez:is-friend-of
ZADD myindex 0 pso:is-friend-of:antirez:matteocollina
ZADD myindex 0 pos:is-friend-of:matteocollina:antirez
```

Теперь все становится интереснее, и я могу запрашивать граф разными способами. Например, с кем `antirez` _дружат_ все люди ?

```
ZRANGE myindex "[spo:antirez:is-friend-of:" "[spo:antirez:is-friend-of:\xff" BYLEX
1) "spo:antirez:is-friend-of:matteocollina"
2) "spo:antirez:is-friend-of:wonderwoman"
3) "spo:antirez:is-friend-of:spiderman"
```

Или какие все отношения `antirez`и `matteocollina`имеют, где первое является субъектом, а второе является объектом?

```
ZRANGE myindex "[sop:antirez:matteocollina:" "[sop:antirez:matteocollina:\xff" BYLEX
1) "sop:antirez:matteocollina:is-friend-of"
2) "sop:antirez:matteocollina:was-at-conference-with"
3) "sop:antirez:matteocollina:talked-with"
```

Комбинируя разные запросы, я могу задавать причудливые вопросы. Например: _Кто все мои друзья, которые, как и пиво, живут в Барселоне, и маттеоколлину тоже считают друзьями? _Чтобы получить эту информацию, я начинаю с `spo`запроса, чтобы найти всех людей, с которыми я дружу. Затем для каждого полученного результата я выполняю `spo`запрос, чтобы проверить, нравится ли им пиво, удаляя те, для которых я не могу найти эту связь. Я делаю это снова, чтобы отфильтровать по городу. Наконец, я выполняю `ops` запрос, чтобы найти из полученного списка, кого Маттеоколлина считает другом.

Обязательно посмотрите [слайды Маттео Коллины о Levelgraph,](http://nodejsconfit.levelgraph.io/) чтобы лучше понять эти идеи.

# Многомерные индексы

Более сложный тип индекса — это индекс, который позволяет выполнять запросы, в которых одновременно запрашиваются две или более переменных для определенных диапазонов. Например, у меня может быть набор данных, представляющий возраст и зарплату людей, и я хочу получить всех людей в возрасте от 50 до 55 лет с зарплатой от 70000 до 85000.

Этот запрос может быть выполнен с индексом из нескольких столбцов, но для этого нам нужно выбрать первую переменную, а затем просмотреть вторую, что означает, что мы можем выполнить гораздо больше работы, чем необходимо. Такие запросы можно выполнять с несколькими переменными, используя разные структуры данных. Например, иногда используются многомерные деревья, такие как _деревья kd_ или _r-деревья . _Здесь мы опишем другой способ индексации данных в нескольких измерениях, используя прием представления, который позволяет нам выполнять запрос очень эффективным способом, используя лексикографические диапазоны Redis.

Допустим, у нас есть точки в пространстве, которые представляют наши выборки данных, где `x`и `y`— наши координаты. Максимальное значение обеих переменных равно 400.

На следующем рисунке синяя рамка представляет наш запрос. Нам нужны все точки `x`от 50 до 100 и `y`от 100 до 300.

![image](https://user-images.githubusercontent.com/3950155/223487818-a23e3514-3f00-44ce-b78e-2795dc826e5f.png)


Чтобы представить данные, которые ускоряют выполнение таких запросов, мы начинаем с дополнения наших чисел 0. Например, представьте, что мы хотим добавить точку 10,25 (x, y) к нашему индексу. Учитывая, что максимальный диапазон в примере равен 400, мы можем просто дополнить до трех цифр, так что мы получим:

```
x = 010
y = 025
```

Теперь мы чередуем цифры, беря крайнюю левую цифру в x, самую левую цифру в y и так далее, чтобы создать одно число:

```
001205
```

Это наш индекс, однако, чтобы легче восстановить исходное представление, если мы хотим (за счет места), мы также можем добавить исходные значения в качестве дополнительных столбцов:

```
001205:10:25
```

Теперь давайте поговорим об этом представлении и о том, почему оно полезно в контексте запросов диапазона. Например, возьмем центр нашего синего прямоугольника, который находится в точке `x=75`и `y=200`. Мы можем закодировать это число, как мы делали ранее, чередуя цифры, получая:

```
027050
```

Что произойдет, если мы заменим две последние цифры соответственно на 00 и 99? Получаем лексикографически непрерывный диапазон:

```
027000 to 027099
```

Это соответствует квадрату, представляющему все значения, где переменная `x` находится в диапазоне от 70 до 79, а `y`переменная находится в диапазоне от 200 до 209. Чтобы определить эту конкретную область, мы можем записать случайные точки в этом интервале.

![image](https://user-images.githubusercontent.com/3950155/223487936-c9c917a4-3b66-4cf5-9672-26f8c069c835.png)

Таким образом, приведенный выше лексикографический запрос позволяет нам легко запрашивать точки в определенном квадрате на изображении. Однако квадрат может быть слишком мал для поля, которое мы ищем, поэтому требуется слишком много запросов. Таким образом, мы можем сделать то же самое, но вместо замены двух последних цифр на 00 и 99 мы можем сделать это для последних четырех цифр, получив следующий диапазон:

```
020000 029999
```

На этот раз диапазон представляет собой все точки в диапазоне `x`от 0 до 99 и `y`от 200 до 299. Рисование случайных точек в этом интервале показывает нам эту большую область.

![image](https://user-images.githubusercontent.com/3950155/223488012-8787bf16-5ed7-4bd4-a0f6-6f2ad6d66e05.png)

Так что теперь наша область слишком велика для нашего запроса, и все еще наше окно поиска не полностью включено. Нам нужно больше детализации, но мы можем легко получить ее, представив наши числа в двоичной форме. На этот раз, когда мы заменяем цифры вместо квадратов, которые в десять раз больше, мы получаем квадраты, которые просто в два раза больше.

Наши числа в двоичной форме, предполагая, что нам нужно всего 9 бит для каждой переменной (чтобы представить числа до 400 в значении), будут:

```
x = 75  -> 001001011
y = 200 -> 011001000
```

Таким образом, при чередовании цифр наше представление в индексе будет таким:

```
000111000011001010:75:200
```

Давайте посмотрим, каковы наши диапазоны, когда мы заменяем последние 2, 4, 6, 8, ... биты на 0 и 1 в представлении с чередованием:

```
2 bits: x between 74 and 75, y between 200 and 201 (range=2)
4 bits: x between 72 and 75, y between 200 and 203 (range=4)
6 bits: x between 72 and 79, y between 200 and 207 (range=8)
8 bits: x between 64 and 79, y between 192 and 207 (range=16)
```

И так далее. Теперь у нас определенно лучшая детализация! Как видите, подстановка N битов из индекса дает нам поля поиска со стороной `2^(N/2)`.

Итак, что мы делаем, так это проверяем измерение, в котором окно поиска меньше, и проверяем ближайшую степень двойки к этому числу. Наше поле поиска было от 50 100 до 100 300, поэтому оно имеет ширину 50 и высоту 200. Мы берем меньшее из двух, 50, и проверяем ближайшую степень двойки, которая равна 64. 64 равно 2^6, поэтому мы будет работать с полученными индексами, заменяющими последние 12 бит из представления с чередованием (так что мы закончим заменой только 6 битов каждой переменной).

Однако отдельные квадраты могут не охватывать весь наш поиск, поэтому нам может понадобиться больше. Что мы делаем, так это начинаем с левого нижнего угла нашего окна поиска, который равен 50 100, и находим первый диапазон, заменяя последние 6 битов в каждом числе на 0. Затем мы делаем то же самое с правым верхним углом.

С двумя тривиальными вложенными циклами for, в которых мы увеличиваем только значащие биты, мы можем найти все квадраты между этими двумя. Для каждого квадрата мы конвертируем два числа в наше чередующееся представление и создаем диапазон, используя преобразованное представление в качестве нашего начала и то же представление, но с включенными последними 12 битами в качестве конечного диапазона.

Для каждого найденного квадрата мы выполняем наш запрос и получаем элементы внутри, удаляя элементы, которые находятся за пределами нашего окна поиска.

Превратить это в код просто. Вот пример Руби:

```
def spacequery(x0,y0,x1,y1,exp)
    bits=exp*2
    x_start = x0/(2**exp)
    x_end = x1/(2**exp)
    y_start = y0/(2**exp)
    y_end = y1/(2**exp)
    (x_start..x_end).each{|x|
        (y_start..y_end).each{|y|
            x_range_start = x*(2**exp)
            x_range_end = x_range_start | ((2**exp)-1)
            y_range_start = y*(2**exp)
            y_range_end = y_range_start | ((2**exp)-1)
            puts "#{x},#{y} x from #{x_range_start} to #{x_range_end}, y from #{y_range_start} to #{y_range_end}"

            # Turn it into interleaved form for ZRANGE query.
            # We assume we need 9 bits for each integer, so the final
            # interleaved representation will be 18 bits.
            xbin = x_range_start.to_s(2).rjust(9,'0')
            ybin = y_range_start.to_s(2).rjust(9,'0')
            s = xbin.split("").zip(ybin.split("")).flatten.compact.join("")
            # Now that we have the start of the range, calculate the end
            # by replacing the specified number of bits from 0 to 1.
            e = s[0..-(bits+1)]+("1"*bits)
            puts "ZRANGE myindex [#{s} [#{e} BYLEX"
        }
    }
end

spacequery(50,100,100,300,6)
```

Хотя это и нетривиально, это очень полезная стратегия индексации, которая в будущем может быть реализована в Redis нативным образом. На данный момент хорошо то, что сложность может быть легко инкапсулирована внутри библиотеки, которую можно использовать для выполнения индексации и запросов. Одним из примеров такой библиотеки является [Redimension](https://github.com/antirez/redimension) , экспериментальная библиотека Ruby, которая индексирует N-мерные данные внутри Redis, используя технику, описанную здесь.

# Многомерные индексы с отрицательными числами или числами с плавающей запятой

Самый простой способ представить отрицательные значения — просто работать с целыми числами без знака и представлять их с использованием смещения, так что при индексировании перед преобразованием чисел в индексированное представление вы добавляете абсолютное значение вашего меньшего отрицательного целого числа.

Для чисел с плавающей запятой самый простой подход, вероятно, состоит в том, чтобы преобразовать их в целые числа, умножив целое число на степень десяти, пропорциональную количеству цифр после точки, которую вы хотите сохранить.

# Недиапазонные индексы

До сих пор мы проверяли индексы, которые полезны для запросов по диапазону или по одному элементу. Однако другие структуры данных Redis, такие как наборы или списки, могут использоваться для создания других типов индексов. Они очень часто используются, но, возможно, мы не всегда понимаем, что на самом деле они являются формой индексации.

Например, я могу индексировать идентификаторы объектов в тип данных Set, чтобы использовать операцию _получения случайных элементов_ через [`SRANDMEMBER`](/commands/srandmember)для получения набора случайных объектов. Наборы также можно использовать для проверки существования, когда все, что мне нужно, это проверить, существует ли данный элемент или нет, имеет ли он одно логическое свойство или нет.

Точно так же списки можно использовать для индексации элементов в фиксированном порядке. Я могу добавить все свои элементы в список Redis и повернуть список, [`RPOPLPUSH`](/commands/rpoplpush)используя то же имя ключа, что и источник и место назначения. Это полезно, когда я хочу обрабатывать заданный набор элементов снова и снова в одном и том же порядке. Подумайте о системе RSS-каналов, которая должна периодически обновлять локальную копию.

Другим популярным индексом, часто используемым с Redis, является **ограниченный список** , в котором элементы добавляются [`LPUSH`](/commands/lpush)и обрезаются с помощью [`LTRIM`](/commands/ltrim), чтобы создать представление только с последними найденными элементами в том же порядке, в котором они были просмотрены.

# Несоответствие индекса

Поддержание индекса в актуальном состоянии может быть сложной задачей, в течение нескольких месяцев или лет возможно появление несоответствий из-за программных ошибок, сетевых разделов или других событий.

Можно использовать разные стратегии. Если данные индекса находятся за пределами Redis, _восстановление чтения_ может быть решением, когда данные исправляются ленивым способом, когда они запрашиваются. Когда мы индексируем данные, которые хранятся в самом Redis, [`SCAN`](/commands/scan)можно использовать семейство команд для постепенной проверки, обновления или перестроения индекса с нуля.
