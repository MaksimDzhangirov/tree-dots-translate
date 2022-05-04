# Шаблон проектирования репозиторий: безболезненный способ упростить логику Вашего Go сервиса

[Данная статья является переводом. Оригинал можно найти по ссылке](https://threedots.tech/post/repository-pattern-in-go/)

Роберт Лащак. Главный инженер [Karhoo](https://www.karhoo.com/). Соучредитель
[Three Dots Labs](https://threedotslabs.com/).
Создатель [Watermill](https://github.com/ThreeDotsLabs/watermill).

За свою жизнь я видел много сложного кода. Довольно часто причиной такой
сложности была логика приложения в сочетании с логикой базы данных. **
Объединение логики вашего приложения с логикой вашей базы данных делает ваше
приложение намного более сложным, трудным для тестирования и поддержки.**

Уже существует проверенный и простой шаблон, решающий эти проблемы. Шаблон,
позволяющий **отделить логику приложения от логики базы данных**. Это позволяет
**упростить код и добавление новых функций**. В качестве бонуса вы можете **
отложить важное решение** по выбору базы данных и её схемы. Еще один хороший
побочный эффект этого подхода — **независимость от поставщика базы данных**. Я
имею в виду шаблона проектирования _репозиторий_.

Когда я вспоминаю приложения, с которыми работал, помню, что было сложно понять,
как они работают. **Я всегда боялся что-то там менять — никогда не знаешь к
каким неожиданным, плохим побочным эффектам это могло привести.** Трудно понять
логику приложения, когда оно смешано с реализацией базы данных. Это также
источник дублирования.

Некоторым спасением здесь может стать
создание [сквозных тестов](https://martinfowler.com/articles/microservice-testing/#testing-end-to-end-introduction)
. Но это скрывает проблему, а не решает её. Проводить большое число сквозных
тестов - это медленное, нестабильное и сложное в обслуживании решение. Иногда
они даже мешают нам создавать новый функционал, а не помогают.

В этой статье я научу вас применять этот шаблон в Go прагматичным, элегантным и
простым способом. Я также подробно рассмотрю тему, которую часто пропускают, -
**искусную обработку транзакций**.

Для этого я подготовил три реализации: **Firestore, MySQL и простую реализацию в
памяти.**

Не вдаваясь в подробности, давайте перейдем к практическим примерам!

> Это не просто очередная статья со случайными фрагментами кода.
>
> Этот пост является частью большого цикла, показывающий как создавать приложения на
> **Go, которые легко разрабатывать, поддерживать и с ними интересно работать в
> долгосрочной перспективе**. Мы делаем это, делясь проверенными методами, основанными
> на многих экспериментах, проведёнными с возглавляемыми нами с командами,
> и [научных исследованиях](https://threedots.tech/post/ddd-lite-in-go-introduction/#thats-great-but-do-you-have-any-evidence-it-works).
> Вы можете изучить эти методы, создав с нами [полнофункциональный](https://threedots.tech/post/serverless-cloud-run-firebase-modern-go-application/#what-wild-workouts-can-do) пример
> веб-приложения на Go - **Wild Workouts**.
>
> Мы поступили не совсем обычно — **добавили некоторые не сразу заметные проблемы
> в первоначальную реализацию Wild Workouts**. Неужели мы сошли с ума? Пока нет. 😉
> Эти проблемы характерны для многих проектов Go. **В долгосрочной перспективе эти
> небольшие проблемы становятся критичными и не позволяют добавлять новые функционал.**
>
> **Это один из важнейших навыков старшего или ведущего разработчика; всегда нужно
> помнить о долгосрочных последствиях.**
>
> Мы исправим их путем **рефакторинга** Wild Workouts. Таким образом, вы быстро поймёте
> методики, которыми мы делимся.
> Знаете ли вы это чувство, когда прочитали статью о какой-то методике и попытались
> реализовать её, но не смогли из-за упущений и пропуска деталей в руководстве.
> Пропуск деталей делает статьи короче и увеличивает просмотры страниц, но это
> не наша цель. Наша цель - создать материал, который даст достаточно знаний для
> применения представленных методик. Если вы еще не читали [предыдущие статьи из
> этого цикла](https://threedots.tech/series/modern-business-software-in-go/),
> мы настоятельно рекомендуем это сделать.
>
> Мы считаем, что в некоторых областях знаний нельзя пропускать детали. Если вы
> хотите быстро и эффективно создавать сложные приложения, вам нужно потратить
> некоторое время, чтобы узнать как это сделать. Если бы всё было так просто,
> у нас бы не было такого большого количества устрашающего устаревшего кода.
>
> Вот [полный список из 14 опубликованных на данный момент статей](https://threedots.tech/series/modern-business-software-in-go/).
>
> Весь код Wild Workouts доступен на [GitHub](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example). Не забудьте поставить звезду
> нашему проекту! ⭐

## Интерфейс репозитория

Идея использования шаблона проектирования репозитория заключается в следующем:

**Давайте абстрагируемся от нашей реализации баз данных, определив
взаимодействие с ней через интерфейс. Вы должны иметь возможность использовать
этот интерфейс для любой реализации базы данных — это означает, что он не должен
содержать каких-либо деталей реализации какой-либо базы данных.**

Начнем с рефакторинга
сервиса [trainer](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/tree/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer)
. В настоящее время сервис позволяет нам получать информацию о доступности часа
для
тренировки [через HTTP API](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/http.go#L17)
и [через gRPC](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/grpc.go#L75)
. Мы также можем изменить доступность часа
[через HTTP API](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/http.go#L12)
и
[gRPC](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/grpc.go#L16)
.

В
предыдущей [статье](https://threedots.tech/post/ddd-lite-in-go-introduction#refactoring-of-trainer-service)
мы рефакторили `Hour`, чтобы в нём использовался упрощенный DDD подход.
Благодаря этому нам не нужно думать о соблюдении правил, когда можно
обновлять `Hour`. Наш уровень предметной области гарантирует, что мы не сможем
сделать ничего "глупого". Нам также не нужно думать ни о какой валидации. Мы
можем просто использовать тип и выполнять необходимые операции.

Нам нужно иметь возможность получить текущее состояние `Hour` из базы данных и
сохранить его. Кроме того, в случае, когда два человека хотят запланировать
тренировку одновременно, только у одного из них должно получиться сделать это
для конкретного часа.

Давайте отразим наши требования в интерфейсе:

```go
package hour

type Repository interface {
	GetOrCreateHour(ctx context.Context, hourTime time.Time) (*Hour, error)
	UpdateHour(
		ctx ctx context.Context,
		hourTime time.Time,
		updateFn func(h *Hour) (*Hour, error),
	) error
}
```

Весь исходный
код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/domain/hour/repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/domain/hour/repository.go#L8)

Мы будем использовать `GetOrCreateHour` для получения данных и `UpdateHour` для
их сохранения.

Мы определяем интерфейс в том же пакете, что и тип `Hour`. Благодаря этому мы
можем избежать дублирования при использовании этого интерфейса во многих
модулях (по моему опыту, это может быть часто). Это также шаблон, аналогичный
`io.Writer`, где пакет `io` определяет интерфейс, а все реализации разделены на
отдельные пакеты.

Как реализовать этот интерфейс?

## Считываем данные

Большинство драйверов баз данных могут использовать `ctx` `context.Context` для
отмены запроса, трассировки и т.д. Он не привязан к какой-либо базе данных
(это обычная концепция Go), поэтому вам не следует бояться, что он каким-то
образом повлияет на предметную область.

В большинстве случаев мы запрашиваем данные, используя UUID или ID, а не
`time.Time`. В нашем случае это нормально — каждый час уникален, исходя из
требований проекта. Я могу представить себе ситуацию, когда у нас было бы
несколько тренеров - в этом случае это предположение будет неверным.
Изменить `time.Time`
на UUID/ID все равно будет просто. А
пока, [YAGNI](https://en.wikipedia.org/wiki/You_aren%27t_gonna_need_it)!

Для наглядности — вот так может выглядеть интерфейс, использующий UUID:

```go
GetOrCreateHour(ctx context.Context, hourUUID string) (*Hour, error)
```

> Вы можете найти пример репозитория на основе UUID в статье [Объединяем DDD, CQRS и
> чистую архитектуру](https://threedots.tech/post/ddd-cqrs-clean-architecture-combined/#repository-refactoring)

Как интерфейс используется в приложении?

```go
import (
// ...
"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part6/internal/trainer/domain/hour"
// ...
)

type GrpcServer struct {
trainer.TrainerServiceServer
hourRepository hour.Repository
}

// ...

func (g GrpcServer) IsHourAvailable(ctx context.Context, request *trainer.IsHourAvailableRequest) (*trainer.IsHourAvailableResponse, error) {
trainingTime, err := protoTimestampToTime(request.Time)
if err != nil {
return nil, status.Error(codes.InvalidArgument, "unable to parse time")
}

h, err := g.hourRepository.GetOrCreateHour(ctx, trainingTime)
if err != nil {
return nil, status.Error(codes.Internal, err.Error())
}

return &trainer.IsHourAvailableResponse{IsAvailable: h.IsAvailable()}, nil
}
```

Весь исходный
код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/grpc.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/0249977c58a310d343ca2237c201b9ba016b148e/internal/trainer/grpc.go#L75)

Никаких сложных вычислений не требуется! Получаем `hour.Hour` и проверяем
доступен ли он. Сможете угадать, какую базу данных мы используем? Нет, в том-то
и дело!

Как я уже упоминал, мы можем избежать привязки к базе и иметь возможность легко
изменить её. Если вы можете поменять базу данных, **это признак того, что вы
правильно реализовали шаблон проектирования репозиторий.** На практике ситуация,
когда вы меняете базу данных, бывает редко. 😉 В случае, если вы используете
решение, которое не размещается на собственном хостинге (например, Firestore),
более важно снизить риск и избежать привязки к поставщику услуг.

Полезным побочным эффектом этого является то, что мы можем отложить решение о
том, какую реализацию базы данных мы хотели бы использовать. Я называю этот
подход _Domain First (Предметная область на первом месте)_. Я подробно описал
его [в предыдущей статье](https://threedots.tech/post/ddd-lite-in-go-introduction/#domain-first-approach)
.
**Если отложить решение о базе данных на потом, можно сэкономить время в начале
проекта. Имея больше информации и контекста, мы также сможем принять лучшее
решение.**

Когда мы используем подход Domain First, первой и самой простой реализацией
репозитория может быть реализация в памяти.

## Пример реализации в памяти

В нашем примере используется простая карта под капотом. `getOrCreateHour`
содержит 5 строк (не считая комментария и одной пустой строки)! 😉

```go
type MemoryHourRepository struct {
hours map[time.Time]hour.Hour
lock  *sync.RWMutex

hourFactory hour.Factory
}

func NewMemoryHourRepository(hourFactory hour.Factory) *MemoryHourRepository {
if hourFactory.IsZero() {
panic("missing hourFactory")
}

return &MemoryHourRepository{
hours:       map[time.Time]hour.Hour{},
lock:        &sync.RWMutex{},
hourFactory: hourFactory,
}
}

func (m MemoryHourRepository) GetOrCreateHour(_ context.Context, hourTime time.Time) (*hour.Hour, error) {
m.lock.RLock()
defer m.lock.RUnlock()

return m.getOrCreateHour(hourTime)
}

func (m MemoryHourRepository) getOrCreateHour(hourTime time.Time) (*hour.Hour, error) {
currentHour, ok := m.hours[hourTime]
if !ok {
return m.hourFactory.NewNotAvailableHour(hourTime)
}

// мы храним часы не как указатели, а как значения
// благодаря этому, мы уверены, что никто не сможет изменить Hour, не используя UpdateHour
return &currentHour, nil
}
```

Весь исходный
код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_memory_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_memory_repository.go#L11)

К сожалению, у реализации памяти есть и недостатки. Самый большой из них
заключается в том, что он не сохраняет данные сервиса после перезапуска. 😉
Этого может быть достаточно для работающей пре-альфа версии. Чтобы наше
приложение было готово к запуску на продакшене, нам нужно какое-то постоянное
хранилище.

## Пример MySQL реализации

Мы уже знаем, как выглядит наша модель и как она себя ведет. Исходя из этого,
давайте определим нашу SQL схему.

```mysql
CREATE TABLE `hours`
(
    hour         TIMESTAMP                                                 NOT NULL,
    availability ENUM ('available', 'not_available', 'training_scheduled') NOT NULL,
    PRIMARY KEY (hour)
);
```

Весь исходный
код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/sql/schema.sql](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/sql/schema.sql#L1)

Когда я работаю с базами данных SQL, я обычно выбираю:

* [sqlx](https://github.com/jmoiron/sqlx) - для простых моделей данных он
  предоставляет полезные функции, которые помогают использовать структуры для
  преобразования результатов запроса. Когда схема усложняется из-за отношений и
  нескольких моделей, пора использовать…
* [SQLBoiler](https://github.com/volatiletech/sqlboiler) - отлично подходит для
  более сложных моделей с множеством полей и отношений, он основан на генерации
  кода. Благодаря этому это происходит очень быстро, и вам не нужно бояться, что
  вы передали неверный `interface{}` вместо другого `interface{}`. 😉
  Сгенерированный код основан на SQL схеме, поэтому вы можете избежать большого
  количества дублирования.

В настоящее время у нас только одна таблица. `sqlx` будет более чем достаточно
😉. Давайте опишем нашу модель БД с учётом «типом хранения».

```go
type mysqlHour struct {
    ID           string    `db:"id"`
    Hour         time.Time `db:"hour"`
    Availability string    `db:"availability"`
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_mysql_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_mysql_repository.go#L17)

> Вы можете спросить, почему бы не добавить атрибут `db` в` hour.Hour`? По
> моему опыту, лучше полностью отделить типы предметных областей от базы данных.
> Его легче тестировать, мы не дублируем проверку и это не приводит к дублированию
> кода. В случае каких-либо изменений в схеме мы можем сделать это только в
> нашей реализации репозитория, а не в половине проекта. Милош описал похожий
> случай в статье [Что нужно знать о DRY](https://threedots.tech/post/things-to-know-about-dry/).
> Я также подробно описал его в [предыдущей статье об упрощенном DDD](https://threedots.tech/post/ddd-lite-in-go-introduction/#the-third-rule---domain-needs-to-be-database-agnostic).

Как мы можем использовать эту структуру?
```go
// sqlContextGetter - это интерфейс, который предоставляет как стандартное подключение к базе данных, так и использующее транзакции
type sqlContextGetter interface {
    GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (m MySQLHourRepository) GetOrCreateHour(ctx context.Context, time time.Time) (*hour.Hour, error) {
    return m.getOrCreateHour(ctx, m.db, time, false)
}

func (m MySQLHourRepository) getOrCreateHour(
    ctx context.Context,
    db sqlContextGetter,
    hourTime time.Time,
    forUpdate bool,
) (*hour.Hour, error) {
    dbHour := mysqlHour{}

    query := "SELECT * FROM `hours` WHERE `hour` = ?"
    if forUpdate {
        query += " FOR UPDATE"
    }

    err := db.GetContext(ctx, &dbHour, query, hourTime.UTC())
    if errors.Is(err, sql.ErrNoRows) {
        // на самом деле эта дата существует, даже если она не сохраняется
        return m.hourFactory.NewNotAvailableHour(hourTime)
    } else if err != nil {
        return nil, errors.Wrap(err, "unable to get hour from db")
    }

    availability, err := hour.NewAvailabilityFromString(dbHour.Availability)
    if err != nil {
        return nil, err
    }

    domainHour, err := m.hourFactory.UnmarshalHourFromDatabase(dbHour.Hour.Local(), availability)
    if err != nil {
        return nil, err
    }

    return domainHour, nil
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_mysql_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_mysql_repository.go#L40)

Для SQL реализации всё просто, потому что нам не нужно поддерживать обратную 
совместимость. В предыдущих статьях мы использовали Firestore в качестве нашей 
основной базы данных. Подготовим реализацию на основе неё, сохраняя обратную 
совместимость.

## Firestore реализация

Если вы хотите провести рефакторинг устаревшего приложения, абстрагирование 
базы данных может быть хорошей отправной точкой.

Иногда приложения создаются с ориентацией на базы данных. В нашем случае -
это подход, ориентированный на HTTP-ответ 😉 - наши модели баз данных основаны 
на моделях, сгенерированных Swagger. Другими словами — наши модели данных 
основаны на моделях данных Swagger, возвращаемых API. Это мешает нам 
абстрагироваться от базы данных? Конечно, нет! Потребуется всего лишь дополнительный
код для преобразования данных.

**Используя подход Domain-First наша модель базы данных была бы намного лучше, 
как в SQL реализации.** Но будем работать с тем, что есть. Давайте шаг за шагом 
избавимся от этого устаревшего кода. Я чувствую, что CQRS поможет нам в этом. 😉

> На практике перенос данных может быть простым до тех пор, пока никакие другие 
> службы не интегрируются напрямую через базу данных.
> 
> К сожалению, это оптимистичное предположение, когда мы работаем с сервисом с устаревшим
> кодом, сервисом с ориентацией на базу данных или CRUD сервисом...

```go
func (f FirestoreHourRepository) GetOrCreateHour(ctx context.Context, time time.Time) (*hour.Hour, error) {
    date, err := f.getDateDTO(
        // getDateDTO следует использовать как для транзакционного, так и для нетранзакционного запроса,
        // лучший способ в этом случае - использовать замыкание
        func() (doc *firestore.DocumentSnapshot, err error) {
            return f.documentRef(time).Get(ctx)
        },
        time,
    )
    if err != nil {
        return nil, err
    }
  
    hourFromDb, err := f.domainHourFromDateModel(date, time)
    if err != nil {
        return nil, err
    }
  
    return hourFromDb, err
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_firestore_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_firestore_repository.go#L31)

```go
// пока что мы сохраняем обратную совместимость, из-за этого метод немного запутан и слишком сложен
// todo - мы исправим это позднее с помощью CQRS :)
func (f FirestoreHourRepository) domainHourFromDateModel(date Date, hourTime time.Time) (*hour.Hour, error) {
    firebaseHour, found := findHourInDateDTO(date, hourTime)
    if !found {
        // на самом деле эта дата существует, даже если она не сохраняется
        return f.hourFactory.NewNotAvailableHour(hourTime)
    }
  
    availability, err := mapAvailabilityFromDTO(firebaseHour)
    if err != nil {
        return nil, err
    }
  
    return f.hourFactory.UnmarshalHourFromDatabase(firebaseHour.Hour.Local(), availability)
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_firestore_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_firestore_repository.go#L120)

К сожалению, интерфейсы Firebase для транзакционных и нетранзакционных запросов 
не полностью совместимы. Чтобы избежать дублирования, я создал `getDateDTO`, который
позволяет решить эту проблему, передав `getDocumentFn`.

```go
func (f FirestoreHourRepository) getDateDTO(
    getDocumentFn func() (doc *firestore.DocumentSnapshot, err error),
    dateTime time.Time,
) (Date, error) {
    doc, err := getDocumentFn()
    if status.Code(err) == codes.NotFound {
        // на самом деле эта дата существует, даже если она не сохраняется
        return NewEmptyDateDTO(dateTime), nil
    }
    if err != nil {
        return Date{}, err
    }
  
    date := Date{}
    if err := doc.DataTo(&date); err != nil {
        return Date{}, errors.Wrap(err, "unable to unmarshal Date from Firestore")
    }
  
    return date, nil
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_firestore_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_firestore_repository.go#L97)

Даже если понадобится дополнительный код, в этом нет ничего плохо. И, по крайней
мере, его можно легко протестировать.

## Обновляем данные
Как я упоминал ранее, очень важно быть уверенным, что только **один человек может 
запланировать тренировку на конкретный час.** Для этого нам нужно использовать **оптимистичную 
блокировку и транзакции**. Даже если транзакции — довольно распространенный термин, 
давайте убедимся что мы понимаем одно и то же под оптимистичной блокировкой.

> Оптимистическое конкурентное управление предполагает, что многие транзакции 
> могут часто завершаться, не мешая друг другу. Во время выполнения транзакции 
> используют ресурсы данных, не блокируя эти ресурсы. Перед фиксацией каждая 
> транзакция проверяет, что никакая другая транзакция не изменила прочитанные 
> данные. Если проверка выявляет конфликтующие модификации, фиксирующая 
> транзакция откатывается и может быть перезапущена.
> 
> [wikipedia.org](https://en.wikipedia.org/wiki/Optimistic_concurrency_control)

Технически в обработке транзакций нет ничего сложного. Самая большая проблема, 
с которой я столкнулся, заключалась в другом — как управлять транзакциями так,
чтобы они не слишком сильно влияли на оставшуюся часть приложения, не зависели
от реализации, были явными и быстрыми.

Я экспериментировал со многими идеями, такими как передача транзакции через 
`context.Context`, передача транзакции через middleware на базе HTTP/gRPC/сообщений и 
т.д. Все подходы, которые я пробовал, имели много серьезных проблем - в них
было слишком много магии, они были не явными и медленными в некоторых случаях.

В настоящий момент мне больше всего нравится подход, основанный на замыкании, 
передаваемом в параметре `undateFn`.

```go
type Repository interface {
    // ...
    UpdateHour(
        ctx context.Context,
        hourTime time.Time,
        updateFn func(h *Hour) (*Hour, error),
    ) error
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/domain/hour/repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/domain/hour/repository.go#L8)

Основная идея заключается в том, что при запуске `UpdateHour` нам необходимо 
передать `updateFn`, который поможет обновить указанный час.

Итак, на практике за одну транзакцию мы:

* получаем и передаём все параметры для updateFn (h *Hour в нашем случае), основываясь 
  на полученном UUID или любом другом параметре (в нашем случае `hourTime` `time.Time`)
* выполняем замыкание (`updateFn` в нашем случае)
* сохраняем возвращённые значения (`*Hour` в нашем случае, если нужно вернём больше значений)
* выполняем откат в случае ошибки, возвращаемой из замыкания

Как это использовать на практике?

```go
func (g GrpcServer) MakeHourAvailable(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
    trainingTime, err := protoTimestampToTime(request.Time)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "unable to parse time")
    }
  
    if err := g.hourRepository.UpdateHour(ctx, trainingTime, func(h *hour.Hour) (*hour.Hour, error) {
        if err := h.MakeAvailable(); err != nil {
            return nil, err
        }
  
        return h, nil
    }); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
  
    return &trainer.EmptyResponse{}, nil
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/grpc.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/0249977c58a310d343ca2237c201b9ba016b148e/internal/trainer/grpc.go#L20)

Как видите, мы получаем экземпляр Hour из какой-то (неизвестной!) базы данных. 
После этого мы делаем этот час доступным (`Available`). Если все нормально, мы
сохраняем час, возвращая его. В рамках [предыдущей статьи](https://threedots.tech/post/ddd-lite-in-go-introduction/) 
**все проверки были перенесены на уровень предметной области, поэтому здесь мы 
уверены, что не делаем ничего «глупого». Это также сильно упростило этот код.** 

```shell
+func (g GrpcServer) MakeHourAvailable(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
@ ...
-func (g GrpcServer) UpdateHour(ctx context.Context, req *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
-	trainingTime, err := grpcTimestampToTime(req.Time)
-	if err != nil {
-		return nil, status.Error(codes.InvalidArgument, "unable to parse time")
-	}
-
-	date, err := g.db.DateModel(ctx, trainingTime)
-	if err != nil {
-		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get data model: %s", err))
-	}
-
-	hour, found := date.FindHourInDate(trainingTime)
-	if !found {
-		return nil, status.Error(codes.NotFound, fmt.Sprintf("%s hour not found in schedule", trainingTime))
-	}
-
-	if req.HasTrainingScheduled && !hour.Available {
-		return nil, status.Error(codes.FailedPrecondition, "hour is not available for training")
-	}
-
-	if req.Available && req.HasTrainingScheduled {
-		return nil, status.Error(codes.FailedPrecondition, "cannot set hour as available when it have training scheduled")
-	}
-	if !req.Available && !req.HasTrainingScheduled {
-		return nil, status.Error(codes.FailedPrecondition, "cannot set hour as unavailable when it have no training scheduled")
-	}
-	hour.Available = req.Available
-
-	if hour.HasTrainingScheduled && hour.HasTrainingScheduled == req.HasTrainingScheduled {
-		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("hour HasTrainingScheduled is already %t", hour.HasTrainingScheduled))
-	}
-
-	hour.HasTrainingScheduled = req.HasTrainingScheduled
-	if err := g.db.SaveModel(ctx, date); err != nil {
-		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to save date: %s", err))
-	}
-
-	return &trainer.EmptyResponse{}, nil
-}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/commit/0249977c58a310d343ca2237c201b9ba016b148e#diff-5e57cb39050b6e252711befcf6fb0a89L20](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/commit/0249977c58a310d343ca2237c201b9ba016b148e#diff-5e57cb39050b6e252711befcf6fb0a89L20)

В нашем случае из `updateFn` мы возвращаем только `(*Hour, error)` - **вы можете
вернуть больше значений, если нужно.** Вы можете возвращать события, модели для 
чтения и т. д.

Теоретически мы также можем использовать тот же `hour.*Hour`, который передали в
`updateFn`. Я решил этого не делать. Использование возвращаемого значения дает 
нам большую гибкость (мы можем заменить его на другой экземпляр `hour.*Hour`, 
если захотим).

Также нет ничего страшного в том, что создать несколько функций, подобных 
`UpdateHour`, с дополнительными данными для сохранения. Под капотом реализация 
должна повторно использовать один и тот же код без большого дублирования.

```go
type Repository interface {
   // ...
   UpdateHour(
      ctx context.Context,
      hourTime time.Time,
      updateFn func(h *Hour) (*Hour, error),
   ) error
  
    UpdateHourWithMagic(
      ctx context.Context,
      hourTime time.Time,
      updateFn func(h *Hour) (*Hour, *Magic, error),
   ) error
}
```
Как это реализовать теперь?

### Реализация транзакций в памяти

Реализация в памяти снова самая простая. 😉 Нам нужно получить текущее значение,
выполнить замыкание и сохранить результат.

Важно что в карте мы храним копию вместо указателя. Благодаря этому мы уверены, 
что без «фиксации» (`m.hours[hourTime] = *updatedHour`) наши значения не 
сохранятся. Ещё раз убедимся в этом, запустив тесты.

```go
func (m MemoryHourRepository) UpdateHour(
    _ context.Context,
    hourTime time.Time,
    updateFn func(h *hour.Hour) (*hour.Hour, error),
) error {
    m.lock.Lock()
    defer m.lock.Unlock()
  
    currentHour, err := m.getOrCreateHour(hourTime)
    if err != nil {
        return err
    }
  
    updatedHour, err := updateFn(currentHour)
    if err != nil {
        return err
    }
  
    m.hours[hourTime] = *updatedHour
  
    return nil
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_memory_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_memory_repository.go#L48)

### Реализация транзакций в Firestore

Реализация в Firestore немного сложнее, но опять же — это связано с обратной совместимостью.
Функции `getDateDTO`, `domainHourFromDateDTO`, `updateHourInDataDTO`, вероятно, 
оказались бы не нужны, когда наша модель данных стала бы лучше. Еще одна причина 
не использовать подход, ориентированный на базу данных/ответ от сервера!

```go
func (f FirestoreHourRepository) UpdateHour(
    ctx context.Context,
    hourTime time.Time,
    updateFn func(h *hour.Hour) (*hour.Hour, error),
) error {
    err := f.firestoreClient.RunTransaction(ctx, func(ctx context.Context, transaction *firestore.Transaction) error {
        dateDocRef := f.documentRef(hourTime)
  
        firebaseDate, err := f.getDateDTO(
            // getDateDTO следует использовать как для транзакционного, так и для нетранзакционного запроса,
            // лучший способ в этом случае - использовать замыкание
            func() (doc *firestore.DocumentSnapshot, err error) {
                return transaction.Get(dateDocRef)
            },
            hourTime,
        )
        if err != nil {
            return err
        }
  
        hourFromDB, err := f.domainHourFromDateModel(firebaseDate, hourTime)
        if err != nil {
            return err
        }
  
        updatedHour, err := updateFn(hourFromDB)
        if err != nil {
            return errors.Wrap(err, "unable to update hour")
        }
        updateHourInDataDTO(updatedHour, &firebaseDate)
  
        return transaction.Set(dateDocRef, firebaseDate)
    })
  
    return errors.Wrap(err, "firestore transaction failed")
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_firestore_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_firestore_repository.go#L52)

Как видите, мы получаем `*hour.Hour`, вызываем `updateFn` и сохраняем результаты 
внутри `RunTransaction`.

**Несмотря на некоторую дополнительную сложность, эта реализация остается чёткой, 
простой для понимания и тестирования.**

### Реализация транзакций в MySQL

Давайте сравним её с реализацией MySQL, где мы лучше спроектировали модели.
Даже если реализация похожа, самая большая разница — это способ обработки 
транзакций.

В SQL драйвере транзакция представлена как `*db.Tx`. Мы используем этот 
конкретный объект для осуществления всех запросов и выполнения отката и фиксации.
Чтобы гарантировать, что мы не забудем о закрытии транзакции, мы выполняем 
откат и фиксацию в `defer`.

```go
func (m MySQLHourRepository) UpdateHour(
    ctx context.Context,
    hourTime time.Time,
    updateFn func(h *hour.Hour) (*hour.Hour, error),
) (err error) {
    tx, err := m.db.Beginx()
    if err != nil {
        return errors.Wrap(err, "unable to start transaction")
    }
  
    // Defer в функции выполняется непосредственно перед выходом.
    // Используя defer, мы можем быть уверены, что мы закроем нашу транзакцию соответствующим образом.
    defer func() {
        // В `UpdateHour` мы используем именованный return - `(err error)`.
        // Благодаря этому можно проверить, завершается ли функция с ошибкой.
        //
        // Даже если функция завершается без ошибок, фиксация транзакции может вернуть ошибку.
        // В этом случае мы можем заменить nil на err `err = m.finish ...`.
        err = m.finishTransaction(err, tx)
    }()
  
    existingHour, err := m.getOrCreateHour(ctx, tx, hourTime, true)
    if err != nil {
        return err
    }
  
    updatedHour, err := updateFn(existingHour)
    if err != nil {
        return err
    }
  
    if err := m.upsertHour(tx, updatedHour); err != nil {
        return err
    }
  
    return nil
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_mysql_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_mysql_repository.go#L82)

В этом случае мы также получаем час, передавая `forUpdate == true` в 
`getOrCreateHour`. Этот флаг добавляет в наш запрос оператор `FOR UPDATE`.
Оператор FOR UPDATE очень важен, потому что без него параллельные 
транзакции не смогут изменять час.

> SELECT ... FOR UPDATE
> Для индексированных записей, обнаруженных при поиске, блокирует строки и любые 
> связанные с индексом записи, как если бы вы выполнили оператор UPDATE для 
> этих строк. Другие транзакции заблокированы от обновления этих строк.
> 
> [dev.mysql.com](https://dev.mysql.com/doc/refman/8.0/en/innodb-locking-reads.html)

```go
func (m MySQLHourRepository) getOrCreateHour(
    ctx context.Context,
    db sqlContextGetter,
    hourTime time.Time,
    forUpdate bool,
) (*hour.Hour, error) {
    dbHour := mysqlHour{}
    
    query := "SELECT * FROM `hours` WHERE `hour` = ?"
    if forUpdate {
        query += " FOR UPDATE"
    }
    // ...
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_mysql_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_mysql_repository.go#L48)

Я никогда не сплю спокойно, если у меня нет автоматического теста на такой код.
Посмотрим на него позднее. 😉

`finishTransaction` выполняется при выходе из `UpdateHour`. В случае сбоя фиксации 
или отката мы также можем переопределить возвращенную ошибку.

```go
// finishTransaction откатывает транзакцию, если указана ошибка.
// Если ошибка равна нулю, транзакция фиксируется.
//
// Если откат не удастся, мы используем библиотеку multierr, чтобы добавить ошибку об ошибке отката.
// Если фиксация не удалась, возвращается ошибка фиксации.
func (m MySQLHourRepository) finishTransaction(err error, tx *sqlx.Tx) error {
    if err != nil {
        if rollbackErr := tx.Rollback(); rollbackErr != nil {
            return multierr.Combine(err, rollbackErr)
        }
  
        return err
    } else {
        if commitErr := tx.Commit(); commitErr != nil {
            return errors.Wrap(err, "failed to commit tx")
        }
  
        return nil
    }
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_mysql_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_mysql_repository.go#L149)

```go
// upsertHour обновляет hour, если он уже существует в базе данных.
// Если не существует, он вставляется.
func (m MySQLHourRepository) upsertHour(tx *sqlx.Tx, hourToUpdate *hour.Hour) error {
    updatedDbHour := mysqlHour{
        Hour:         hourToUpdate.Time().UTC(),
        Availability: hourToUpdate.Availability().String(),
    }
  
    _, err := tx.NamedExec(
              `INSERT INTO
                      hours (hour, availability)
               VALUES
                      (:hour, :availability)
               ON DUPLICATE	KEY UPDATE
                      availability = :availability`,
        updatedDbHour,
    )
    if err != nil {
        return errors.Wrap(err, "unable to upsert hour")
    }
  
    return nil
}

```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/hour_mysql_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb/internal/trainer/hour_mysql_repository.go#L122)

## Заключение

Даже если подход с использованием репозитория добавляет немного больше кода, это
полностью себя окупает. **На практике вы можете потратить на это 5 минут и 
вложения вскоре окупятся.**

В этой статье мы упускаем одну важную часть — тесты. Теперь добавлять тесты 
должно быть намного проще, но все ещё может быть неочевидно, как это делать 
правильно.

Чтобы статья не стало очень большой, я расскажу о них в ближайшие 1-2 недели. 🙂
Полностью код с рефактирингом, включая тесты, [доступен на GitHub](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/commit/34c74e9d2cbc80160b4ff26e59818a89d10aa1eb).

Напоминаем, что вы также можете запустить приложение [одной командой](https://threedots.tech/post/serverless-cloud-run-firebase-modern-go-application/#running) и найти 
весь исходный код на [GitHub](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)!

Другая методика, которая довольно хорошо работает — это чистая/гексагональная 
архитектура — Милош описывает её в статье [Введение в чистую архитектуру](https://threedots.tech/post/introducing-clean-architecture).

Считаете ли Вы такую методику полезной для вашего приложения? Используете ли вы уже
шаблон репозиторий как-то по-другому? **Дайте нам знать об этом в комментариях!**