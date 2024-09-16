# Внедряем pg

## Расширить структуру IndexDeclaration для возможности описания условных индексов

 Добавляем:

- срез строк Conditions с условиями индекса
- структуру Order с полями:
  - field string
  - order uint8 "desc|ask"

## Модифицируем SelectorLimiter

Добавляем поле after как срез any. Необходимо для того, что бы делать запросы с limit, after.

## Возможность поднять всю таблицу

При описании класса надо предусмотреть параметр EnableSelectAll который сгенерирует метод позволяющий делать select без условий (необходимо для использования словарей). В октопусе сделать невозможным использование такой настройки.

## Работа селектора

- проверяем, что переданы ключи для выборки если это не selectAll
- делаем проверку на количество переданных ключей (здравый смысл использования оператора IN)
- определяем является ли запрос балковым, либо передано несколько ключей либо запрос по неуникальному индексу
- делаем проверку на то: что указан лимит для неуникального ключа
- важно учитывать, что ключи могут быть в виде среза, когда индекс по нескольким ключам
- если же на поля, по которым делается запрос навешены сериализаторы то сериалезуем ключи для выполнения запроса
- если передан selectorLimiter с after, то необходимо определить порядок сортировки в индексе и выбрать правильное условие > или <
- можно добавить дополнительные условия и сортировку после того: как выбрали необходимое через подзапрос  

## Работа с mock

За основу берём https://github.com/pashagolub/pgxmock он позволяет поднять сервер претворяющийся базой данных и сверху делаем свое поведение мока.

## Метрики

Необходимо переиспользовать метрики с octopus-а. Если это возможно то надо уйти от дублирования шаблонов. Один шаблон для нескольких движков или вынос метрик из шаблонизатора.

## Проверить возможность вызова встроенных процедур

ProcFields для postgres. Дописать доку в этом месте

## Написать сериализаторы для PostgrfeSQL

- intarray