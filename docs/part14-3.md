# Более безопасные перечисления в Go

Милош Смолка. Технический руководитель [Karhoo](https://www.karhoo.com/). Соучредитель
[Three Dots Labs](https://threedotslabs.com/). Создатель [Watermill](https://github.com/ThreeDotsLabs/watermill).

Перечисления являются важной частью веб-приложений. Go не поддерживает их из 
коробки, но есть способы их эмулировать.

Многие очевидные решения далеки от идеальных. Вот некоторые идеи, которые мы 
используем, чтобы сделать перечисления более безопасными.

## iota

Go позволяет перечислять вещи с помощью iota.

```go
const (
    Guest = iota
    Member
    Moderator
    Admin
)
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/01-iota/role/role.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/01-iota/role/role.go#L3)

В то время как в Go всё определено явно, `iota` кажется чем-то магическим. Если вы
отсортируете группу иначе, это приведёт к побочным эффектам. В приведенном выше 
примере, вы можете случайно создать модераторов вместо членов. Вы можете явно 
присвоить номер каждому значению, чтобы избежать этой проблемы, но в этом случае
`iota` не нужна.

`iota` хорошо работает для флагов, представленных степенями двойки.

```go
const (
    Guest = 1 << iota   // 1
    Member              // 2
    Moderator           // 4
    Admin               // 8
)

// ...

user.Role = Member | Moderator // 6
```

Битовые маски эффективны и иногда полезны. Однако это другой вариант 
использования, чем перечисления в большинстве веб-приложений. Часто вам будет 
удобно хранить все роли в списке. Это также будет более читабельно.

Основная проблема с `iota` заключается в том, что она работает с целыми числами, 
которые не защищают от передачи недопустимых значений.

```go
func CreateUser(r int) error {
	fmt.Println("Creating user with role", r)
	return nil
}

func main() {
    err := CreateUser(-1)
    if err != nil {
        fmt.Println(err)
    }

    err = CreateUser(42)
    if err != nil {
        fmt.Println(err)
    }
}
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/01-iota/main.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/01-iota/main.go)

Функция `CreateUser` с радостью примет на вход -1 или 42 даже при отсутствии 
соответствующих ролей.

Конечно, мы могли бы проверить это внутри функции. Но мы используем язык со 
строгими типами, так что давайте воспользуемся этим. В контексте нашего 
приложения роль пользователя — это гораздо больше, чем просто какое-то число.

> Антипаттерн: Целочисленные перечисления
> 
> Не используйте целые числа, основанные на `iota` для представления перечислений, 
> которые не являются последовательными числами или флагами.

Мы могли бы ввести тип для улучшения решения.

```go
type Role uint

const (
    Guest Role = iota
    Member
    Moderator
    Admin
)
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/02-typed-iota/role/role.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/02-typed-iota/role/role.go#L5)

Выглядит лучше, но всё равно вместо `Role` можно передать любое 
произвольное целое число. Компилятор Go нам здесь не поможет.

```go
func CreateUser(r int) error {
	fmt.Println("Creating user with role", r)
	return nil
}

func main() {
    err := CreateUser(-1)
    if err != nil {
        fmt.Println(err)
    }

    err = CreateUser(42)
    if err != nil {
        fmt.Println(err)
    }
}
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/02-typed-iota/main.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/02-typed-iota/main.go)

Тип — это улучшение по сравнению с обычным целым числом, но это все еще иллюзия.
Это не дает нам никаких гарантий, что роль действительна.

## Ограничивающие значения

Поскольку `iota` начинается с нуля, `Guest` - это также `Role` с нулевым 
значением. Это затрудняет определение того, пуста ли роль или кто-то передал 
значение `Guest`.

Этого можно избежать, считая от 1. Еще лучше оставить явное ограничивающее значение,
которое можно сравнить и которое нельзя спутать с реальной ролью.

```go
type Role uint

const (
    Unknown Role = iota
    Guest
    Member
    Moderator
    Admin
)
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/02-typed-iota/role/role.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/02-typed-iota/role/role.go#L6)

```go
func CreateUser(r role.Role) error {
    if r == role.Unknown {
        return errors.New("no role provided")
    }
    
    fmt.Println("Creating user with role", r)
    
    return nil
}
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/02-typed-iota/main.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/02-typed-iota/main.go#L11)

> Тактика: явные ограничивающие значения
> 
> Используйте явную переменную для нулевого значения перечисления.

## Slugs

Кажется, что перечисления должны представлять собой последовательные целые 
числа, но это редко бывает корректным представлением. В веб-приложениях мы 
используем перечисления для группировки возможных вариантов некоторого типа. 
Они плохо соотносятся с числами.

Трудно понять контекст, когда вы видите `3` в ответе API, таблице базы данных 
или логах. Вы должны посмотреть в исходный код или устаревшую документацию, 
чтобы узнать, о чем идет речь.

Строковые имена более значимы, чем целые числа, в большинстве случаев. Везде, 
где вы его видите, `moderator` очевиден. Так как `iota` нам все равно не поможет, 
мы можем также использовать удобочитаемые строки.

```go
type Role string

const (
    Unknown   Role = ""
    Guest     Role = "guest"
    Member    Role = "member"
    Moderator Role = "moderator"
    Admin     Role = "admin"
)
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/03-slugs/role/role.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/03-slugs/role/role.go#L5)

> Тактика: Slugs
> 
> Используйте строковые значения вместо целых чисел.
> 
> Избегайте пробелов для облегчения синтаксического анализа и логирования. 
> Используйте camelCase, snake_case или kebab-case.

Slugs особенно полезны для кодов ошибок. Ответ об ошибке, такой как `{"error":
"user-not-found"}`, очевиден в отличие от `{"error": 4102}`.

Однако этот тип по-прежнему может содержать любую произвольную строку.

```go
err = CreateUser("super-admin")
if err != nil {
    fmt.Println(err)
}
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/03-slugs/main.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/03-slugs/main.go#L36)

## Строковые Enums

Последняя итерация использует структуры. Это позволяет нам работать с кодом,
безопасным по своему способу написания. Нам не нужно проверять правильность 
переданного значения.

```go
type Role struct {
    slug string
}

func (r Role) String() string {
    return r.slug
}

var (
    Unknown   = Role{""}
    Guest     = Role{"guest"}
    Member    = Role{"member"}
    Moderator = Role{"moderator"}
    Admin     = Role{"admin"}
)
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/04-structs/role/role.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/04-structs/role/role.go#L14)

Поскольку поле `slug` не экспортируемое, невозможно заполнить его снаружи пакета. 
Единственная неправильная роль, которую вы можете создать — это пустая: `Role{}`.

Мы можем добавить конструктор для создания корректной роли с 
использованием slug:

```go
func FromString(s string) (Role, error) {
	switch s {
	case Guest.slug:
		return Guest, nil
	case Member.slug:
		return Member, nil
	case Moderator.slug:
		return Moderator, nil
	case Admin.slug:
		return Admin, nil
	}

	return Unknown, errors.New("unknown role: " + s)
}
```
Весь исходный код: [github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/04-structs/role/role.go](https://github.com/ThreeDotsLabs/go-web-app-antipatterns/blob/master/02-enums/04-structs/role/role.go#L21)

> Тактика: Строковые Enums
> 
> Инкапсулируйте перечисления в структуры для дополнительной безопасности во 
> время компиляции.

Этот подход идеален, когда вы работаете с бизнес-логикой. Сохранение структур 
в корректном состоянии в памяти упрощает работу с вашим кодом и его понимание. 
Достаточно проверить, не является ли тип enum пустым, и вы уверены, что это 
корректное значение.

При таком подходе существует одна потенциальная проблема. Структуры не могут быть 
константами в Go, поэтому можно перезаписать глобальные переменные следующим 
образом:

```go
roles.Guest = role.Admin
```

Однако для этого нет разумной причины. Более вероятно, вы случайно передадите 
недопустимое целое число.

Другим недостатком является то, что вы должны обновлять информацию в двух 
местах: в списке перечислений и в конструкторе. Тем не менее, это легко 
заметить, даже если вы сначала пропустите это.

## А как вы решаете эту проблему?

Когда отсутствует какой-то функционал в Go, мы склонны создавать свои собственные
ведения дел. То, что я описал, является лишь одним из возможных подходов. 
Используете ли вы другой шаблон для хранения ваших перечислений? Пожалуйста, 
поделитесь им в комментариях. 🙂

Полный исходный код смотри в [репозитории антишаблонов](https://github.com/ThreeDotsLabs/go-web-app-antipatterns).
