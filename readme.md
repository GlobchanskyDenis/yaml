## Помощник для конфигурации модулей yaml файлами

Модуль читающий yaml файлы заполняющий поблочно конфигурационные dto.

Наличие в dto тега `conf` означает что поле однозначно должно быть задано в конфигурации. В случае отсутствия данного поля в конфигурации будет сгенерирована ошибка. В случае отсутствия тега `conf` в конфигурационнике данное поле не будет заполнено из конфигурационника. Для исчислимых типов допустимо задавать теги `max` и `min`. В случае если значение в конфигурационнике будет нарушать условие тегов `max` или `min` - будет сформирована соответствующая ошибка.

Добавлена поддержка тега `env` для строкового типа. Поле в которое добавлен метатег `env` будет заполнено переменной окружения с именем содержащимся В КОНФИГУРАЦИОННИКЕ. В метатег должно быть записано значение `true`.

Добавлена поддержка тега `enum` для перечислимых и строкового типов. Поле в которое добавлен метатег `env` будет проверено на соответствие одному из предложенных вариантов из метатега. Варианты перечисляются через символ `;`.

> Модуль работает со всеми примитивами данных.

> Модуль работает с комплексными типами данных.

> Модуль не работает с интерфейсами.

## Пример

Код

```
   // Конструктор
   config := NewConfigurator()

   // Читаем файл с конфигурацией различных модулей. Объект сохраняет информацию о последнем считанном файле
   if err := config.ReadFile("sample_test.yaml"); err != nil { /* handle error */ }

   // Конфигурируем по отдельности каждый модуль используя алиасы
   type DtoType struct {
      Host:        string `conf:"Host" env:"true"`
      DatabaseName string `conf:"DBName"`
      ConnAmount   uint   `conf:"ConnAmount" min:"1"`
      SimpleQuery  string `conf:"SimpleQuery" enum:"ON;OFF"`
   }
   var dto DtoType
   if err := config.ParseToStruct(&dto, "Alias1"); err != nil { /* handle error */ }
```

Содержимое конфигурационного файла

```
   Alias1:
      Host: DB_HOST_ENV
      DBName: my_database
      ConnAmount: 3
      SimpleQuery: ON
   Alias2:
      WorkersAmount: 5
```
