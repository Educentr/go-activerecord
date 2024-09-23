# ORM

Схема Active Record — это подход к доступу к данным в базе данных.

Таблица базы данных или представление обёрнуты в классы. Таким образом, объектный экземпляр привязан к единственной строке в таблице. После создания объекта новая строка будет добавляться к таблице на сохранение. Любой загруженный объект получает свою информацию от базы данных. Когда объект обновлён, соответствующая строка в таблице также будет обновлена. Класс обёртки реализует методы средства доступа или свойства для каждого столбца в таблице или представлении.

см. так же:

- [docs/intro.md](https://github.com/Educentr/go-activerecord/blob/main/docs/intro.md)
- [docs/manual.md](https://github.com/Educentr/go-activerecord/blob/main/docs/manual.md)
- [docs/cookbook.md](https://github.com/Educentr/go-activerecord/blob/main/docs/cookbook.md)
