# Безопасный репозиторий по своему строению: крепче спим, не опасаясь уязвимостей связанных с безопасностью

[Данная статья является переводом. Оригинал можно найти по ссылке](https://threedots.tech/post/repository-secure-by-design/)

Роберт Лащак. Главный инженер [Karhoo](https://www.karhoo.com/). Соучредитель
[Three Dots Labs](https://threedotslabs.com/).
Создатель [Watermill](https://github.com/ThreeDotsLabs/watermill).

Благодаря тестам и код ревью вы можете сделать так, чтобы ваш проект не содержал
ошибок. Правильно? Ну… на самом деле, наверное, нет. Это было бы слишком просто. 😉 
Эти методики снижают вероятность ошибок, но не могут полностью их устранить.
Но значит ли это, что нам нужно жить с мыслью о появлении ошибок до конца 
жизни?

Более года назад я нашел довольно [интересный PR](https://github.com/goharbor/harbor/pull/8917/files) в 
проекте `harbor`. Это было исправление проблемы, которая **позволяла создать 
пользователя-администратора обычным пользователям. Очевидно, это была серьезная
проблема безопасности.** Конечно, автоматические тесты не нашли эту ошибку ранее.

Вот как выглядит исправление ошибки:

```shell
		ua.RenderError(http.StatusBadRequest, "register error:"+err.Error())
		return
	}
+
+	if !ua.IsAdmin && user.HasAdminRole {
+		msg := "Non-admin cannot create an admin user."
+		log.Errorf(msg)
+		ua.SendForbiddenError(errors.New(msg))
+		return
+	}
+
	userExist, err := dao.UserExists(user, "username")
	if err != nil {
```

Один оператор `if` исправил ошибку. Добавление новых тестов также должно 
гарантировать отсутствие этой ошибки в будущем. Это достаточно? **Защитило ли это
приложение от подобной ошибки в будущем? Я уверен, что нет.**

Проблема становится ещё больше в более сложных системах, над которыми работает 
большая команда разработчиков. Что, если кто-то новичок в проекте и забудет 
добавить это выражение `if`? Даже если вы не нанимаете новых людей в настоящее 
время, они могут быть наняты в будущем. **Вы, вероятно, удивитесь, как долго 
будет использоваться написанный вами код.** Мы не должны доверять людям 
использовать созданный нами код по назначению — они этого не сделают.

**В некоторых случаях решение, которое защитит нас от подобных проблем, — это 
хорошее проектирование. Хорошее проектирование не должен позволять использовать 
наш код недопустимым образом.** Хорошее проектирование должен гарантировать, что
вы можете без страха прикасаться к существующему коду. Новички в проекте будут 
чувствовать себя в большей безопасности, внося изменения.

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
> не наша цель. Наша цель — создать материал, который даст достаточно знаний для
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

В этой статье я покажу, как я добился того, что только люди, которым это 
разрешено, смогут просматривать и редактировать тренировку. В нашем случае 
тренировку могут видеть только владелец тренировки (участник) и тренер. Я реализую
это таким образом, чтобы не позволить использовать наш код не по назначению.
По своей структуре.

Наше текущее приложение предполагает, что репозиторий — это единственный способ 
доступа к данным. Из-за этого я добавлю авторизацию в слой репозитория. **При этом 
мы уверены, что доступ к этим данным неавторизованным пользователям невозможен.**

> Что такое репозиторий? Короче говоря, если у вас не было возможности прочитать 
> наши предыдущие статьи, репозиторий — это шаблон проектирования, который 
> помогает нам абстрагировать реализацию базы данных от логики нашего приложения.
> Если вы хотите узнать больше о его преимуществах и узнать, как применить его в
> своем проекте, прочитайте мою предыдущую статью: [Шаблон проектирования 
> репозиторий: безболезненный способ упростить логику Вашего Go сервиса](https://threedots.tech/post/repository-pattern-in-go/).

Но подождите, является ли репозиторий подходящим местом для управления 
авторизацией? Ну, я могу себе представить, что некоторые люди могут скептически 
относиться к такому подходу. Конечно, мы можем начать философскую дискуссию о 
том, что может быть в репозитории, а что нет. Кроме того, фактическая логика 
того, кто может видеть тренировку, будет размещена на уровне _предметной области_. 
Существенных минусов не вижу, а плюсы очевидны. На мой взгляд, здесь должен 
победить прагматизм.

> Что еще интересно, в этом цикле мы ориентируемся на _бизнес-ориентированные_ 
> _приложения_. Но даже если проект _Harbour_ является чисто системным приложением, к
> нему также можно применить большинство представленных шаблонов.
>
> Познакомив нашу команду с [чистой архитектурой](https://threedots.tech/post/introducing-clean-architecture/), наш товарищ по команде 
> использовал этот подход в своей игре для абстрактного движка рендеринга. 😉
> 
> (Привет, Мариуш, если ты это читаешь!)

## Покажите мне код, пожалуйста!

Чтобы проектирование было надёжным, нам нужно реализовать три вещи:

1. Логика, кто может видеть тренировку (в слое [предметной области](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/tree/v2.6/internal/trainer/domain/hour)),
2. Функции, используемые для тренировки (`GetTraining` в [репозитории](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/186a2c4a912e485ac7bb4d18c2892df7617e9ec9/internal/trainings/adapters/trainings_firestore_repository.go#L57)),
3. Функции, используемые для обновления тренировки (`UpdateTraining` в [репозитории](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/186a2c4a912e485ac7bb4d18c2892df7617e9ec9/internal/trainings/adapters/trainings_firestore_repository.go#L83)).

### Слой предметной области

Первая часть — это логика, отвечающая за принятие решения о том, может ли 
кто-то увидеть тренировку. Поскольку это часть логики предметной области (вы 
можете обсудить, кто может видеть тренировку с вашей командой или продуктовой 
командой), она должна перейти в слой предметной области. Это реализовано с 
помощью функции `CanUserSeeTraining`.

Также допустимо хранить её в слое предметной области, но её сложнее 
использовать повторно. Я не вижу никакого преимущества в таком подходе — тем 
более, если перевести её в предметную область ничего не стоит. 😉

```go
package training

// ...

type User struct {
    userUUID string
    userType UserType
}

// ...

type ForbiddenToSeeTrainingError struct {
    RequestingUserUUID string
    TrainingOwnerUUID  string
}

func (f ForbiddenToSeeTrainingError) Error() string {
    return fmt.Sprintf(
        "user '%s' can't see user '%s' training",
        f.RequestingUserUUID, f.TrainingOwnerUUID,
    )
}

func CanUserSeeTraining(user User, training Training) error {
    if user.Type() == Trainer {
        return nil
    }
    if user.UUID() == training.userUUID {
        return nil
    }
    
    return ForbiddenToSeeTrainingError{user.UUID(), training.UserUUID()}
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainings/adapters/trainings_firestore_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/186a2c4a912e485ac7bb4d18c2892df7617e9ec9/internal/trainings/adapters/trainings_firestore_repository.go#L57)

### Репозиторий

Теперь, когда у нас есть функция `CanUserSeeTraining`, нам нужно использовать эту
функцию. Всё довольно просто.

```shell
func (r TrainingsFirestoreRepository) GetTraining(
	ctx context.Context,
	trainingUUID string,
+	user training.User,
) (*training.Training, error) {
	firestoreTraining, err := r.trainingsCollection().Doc(trainingUUID).Get(ctx)

	if status.Code(err) == codes.NotFound {
		return nil, training.NotFoundError{trainingUUID}
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get actual docs")
	}

	tr, err := r.unmarshalTraining(firestoreTraining)
	if err != nil {
		return nil, err
	}
+
+	if err := training.CanUserSeeTraining(user, *tr); err != nil {
+		return nil, err
+	}
+
	return tr, nil
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainings/adapters/trainings_firestore_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/186a2c4a912e485ac7bb4d18c2892df7617e9ec9/internal/trainings/adapters/trainings_firestore_repository.go#L57)

Не слишком ли это просто? Наша цель — создать простой, а не сложный проект и
код. **Это отличный признак того, что это очень просто.**

Обновляем UpdateTraining подобным образом.

```shell
func (r TrainingsFirestoreRepository) UpdateTraining(
	ctx context.Context,
	trainingUUID string,
+	user training.User,
	updateFn func(ctx context.Context, tr *training.Training) (*training.Training, error),
) error {
	trainingsCollection := r.trainingsCollection()

	return r.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		documentRef := trainingsCollection.Doc(trainingUUID)

		firestoreTraining, err := tx.Get(documentRef)
		if err != nil {
			return errors.Wrap(err, "unable to get actual docs")
		}

		tr, err := r.unmarshalTraining(firestoreTraining)
		if err != nil {
			return err
		}
+
+		if err := training.CanUserSeeTraining(user, *tr); err != nil {
+			return err
+		}
+
		updatedTraining, err := updateFn(ctx, tr)
		if err != nil {
			return err
		}

		return tx.Set(documentRef, r.marshalTraining(updatedTraining))
	})
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainings/adapters/trainings_firestore_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/186a2c4a912e485ac7bb4d18c2892df7617e9ec9/internal/trainings/adapters/trainings_firestore_repository.go#L83)

И... на этом все! Есть ли способ, чтобы кто-то смог использовать это неправильно? 
Пока `User` находится в валидном состоянии – нет.

Этот подход аналогичен методу, представленному во [вводной статье по упрощенному
DDD](https://threedots.tech/post/ddd-lite-in-go-introduction/). Все дело в создании кода, который мы не можем использовать неправильно.

Вот как теперь выглядит использование `UpdateTraining`:

```go
func (h ApproveTrainingRescheduleHandler) Handle(ctx context.Context, cmd ApproveTrainingReschedule) (err error) {
    defer func() {
        logs.LogCommandExecution("ApproveTrainingReschedule", cmd, err)
    }()
    
    return h.repo.UpdateTraining(
        ctx,
        cmd.TrainingUUID,
        cmd.User,
        func(ctx context.Context, tr *training.Training) (*training.Training, error) {
            originalTrainingTime := tr.Time()
    
            if err := tr.ApproveReschedule(cmd.User.Type()); err != nil {
                return nil, err
            }
    
            err := h.trainerService.MoveTraining(ctx, tr.Time(), originalTrainingTime)
            if err != nil {
                return nil, err
            }
    
            return tr, nil
        },
    )
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainings/app/command/approve_training_reschedule.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/22c0a25b67c4669d612a2fa4a434ffae8e35e65a/internal/trainings/app/command/approve_training_reschedule.go#L39)

Конечно существуют ещё некоторые правила, если тренировка (`Training`) может 
быть перенесена, но они находятся в предметной области `Training`. Подробно 
это описано во [вводной статье по упрощенному DDD](https://threedots.tech/post/ddd-lite-in-go-introduction/). 😉

## Работа с коллекциями

Даже если этот подход идеально подходит для работы с одним тренингом, вы должны 
быть уверены, что доступ к набору тренировок надежно защищен. Здесь нет никакой 
магии:

```go
func (r TrainingsFirestoreRepository) FindTrainingsForUser(ctx context.Context, userUUID string) ([]query.Training, error) {
	query := r.trainingsCollection().Query.
		Where("Time", ">=", time.Now().Add(-time.Hour*24)).
		Where("UserUuid", "==", userUUID).
		Where("Canceled", "==", false)

	iter := query.Documents(ctx)

	return r.trainingModelsToQuery(iter)
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainings/adapters/trainings_firestore_repository.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/186a2c4a912e485ac7bb4d18c2892df7617e9ec9/internal/trainings/adapters/trainings_firestore_repository.go#L182)

Делать это в слое приложения с помощью функции `CanUserSeeTraining` будет 
очень дорого и медленно. Лучше создать небольшое логическое дублирование.

Если эта логика в вашем приложении является более сложной, вы можете 
попытаться абстрагировать её в слое предметной области в формат, который вы 
можете преобразовать в параметры запроса при обращении к вашей базы данных.
Однажды я делал так, и это сработало довольно хорошо.

Но в Wild Workouts это добавит ненужной сложности — давайте 
[Делать проще, тупица](https://en.wikipedia.org/wiki/KISS_principle).

## Обработка внутренних обновлений

Мы часто хотим иметь конечные точки, которые позволяют разработчику или 
операционному отделу вашей компании вносить некоторые изменения "тайно".
Худшее, что вы можете сделать в этом случае, — это создать всякого рода 
«фейковых пользователей» и хаки.

Из моего опыта это заканчивается большим количеством операторов if, добавленных
в код. Это также запутывает журнал аудита (если он у вас есть). Вместо 
«фальшивого пользователя» лучше создать специальную роль и явно определить 
права доступа для роли.

Если нужно работать с методами репозитория, не требующими знания текущего 
пользователя (для обработчиков сообщений в модели Издатель/Подписчик или 
миграций), лучше создать отдельные методы репозитория. В этом случае их название 
имеет важное значение — мы должны быть уверены, что человек, использующий этот 
метод, знает о последствиях для безопасности.

По моему опыту, если обновления для разных акторов сильно различаются, стоит 
даже ввести отдельные [CQRS команды](https://threedots.tech/post/basic-cqrs-in-go/) для каждого актора. В нашем случае это 
может быть `UpdateTrainingByOperations`.

## Передача аутентификации через `context.Context`

Насколько я знаю, некоторые люди передают данные аутентификации через 
`context.Context`.

Я настоятельно рекомендую не передавать ничего, что требуется вашему приложению 
для правильной работы через `context.Context`. Причина проста — при передаче 
значений через `context.Context` мы теряем одно из самых существенных 
преимуществ Go — статическую типизацию. Это также скрывает, что именно является 
вводными параметрами для ваших функций.

Если вам по какой-то причине нужно передавать значения через контекст, это 
может быть признаком плохого проектирования где-то в вашем сервисе. Может, 
функция слишком много делает, и передать туда все аргументы сложно? Может пора 
её разбить?

## И это все на сегодня!

Как видите, представленный подход легко быстро реализовать.

Я надеюсь, что это поможет вам в вашем проекте и придаст вам больше уверенности
в будущих разработках.

Видите ли вы, что это может помочь в вашем проекте? Считаете ли вы, что это 
может помочь вашим коллегам? Не забудьте поделиться с ними!