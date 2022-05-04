# Вы не должны создавать свою собственную систему аутентификации. Пусть Firebase сделает это за вас

[Данная статья является переводом. Оригинал можно найти по ссылке](https://threedots.tech/post/firebase-cloud-run-authentication/)

Роберт Лащак. Главный инженер [Karhoo](https://www.karhoo.com/). Соучредитель
[Three Dots Labs](https://threedotslabs.com/). Создатель [Watermill](https://github.com/ThreeDotsLabs/watermill).

Добро пожаловать в третью и **заключительную статью о том, как создать** 
"Слишком современное Go приложение". Но не волнуйтесь. Это не означает, что мы 
закончили показывать вам, как создавать приложения, которые легко разрабатывать, 
поддерживать и приятно работать в долгосрочной перспективе. **На самом деле это
только начало ещё большей серии!**

Мы намеренно создали текущую версию приложения таким образом, чтобы его было
сложно поддерживать и развивать в будущем. Но сделали это утонченно. 😉 В 
следующей статье мы начнем процесс рефакторинга. **Мы покажем вам, как едва 
заметные и упрощения могут стать проблемой в долгосрочной перспективе. Даже 
добавить новый функционал теперь может быть не так просто как разработать всё 
приложение с этим функционалом с нуля (даже несмотря на то, что приложение было
создано не так давно).**

Но прежде чем приступить у нас есть одна важная тема для обсуждения — 
аутентификация. Это та часть приложения, которая не нуждается в рефакторинге. 😉

## Не стоит изобретать велосипед

(Я надеюсь, что) Вы не создаёте язык программирования и фреймворк для каждого 
проекта, над которым работаете. Даже если Вы делаете так, **кроме потери времени, 
это ничем не навредит. Но это не касается аутентификации. Почему?**

Представим, что вы работаете в компании, внедряющей рынок криптовалют. В начале 
проекта вы решили создать свой собственный механизм аутентификации. Первая 
причина — ваш начальник, который не доверяет внешним сервисам аутентификации.
Вторая причина заключается в том, что вы считаете, что это должно быть просто.

_«Ты делал это много раз»_ - вот что каждый раз повторяют странные голоса в 
твоей голове.

Вы смотрели документальный сериал "Расследования авиакатастроф"? В польской
версии каждая серия начинается примерно со слов: _«Катастрофы - это не 
случайность, а последовательность досадных событий»_. В программировании это 
предложение на удивление верно.

В нашем гипотетическом проекте вначале ваша компания решила ввести сервисный 
аккаунт, который позволял переводить средства между всеми аккаунтами. Даже из 
кошелька, который в теории, может быть не подключен к системе. Конечно, временно. 😆
Звучит не очень, но с самого начала это упростило поддержку клиентов.

Со временем об этой функции все забыли. До тех пор, пока однажды хакеры не 
обнаружили ошибку в механизме аутентификации, которая позволила им взломать любой
аккаунт на нашем рынке. Включая сервисный аккаунт.

Наш босс и сотрудники были не так счастливы, как хакеры, потому что потеряли 
все деньги. Проблема не была бы такой большой, если бы мы потеряли «только» все 
деньги компании. В этом случае мы также потеряли все деньги наших клиентов 😉.

![itsgone](images/part4/aannnd-its-gone.jpg)

Этот пример может показаться крайностью и маловероятным. Действительно ли это не
часто встречающийся случай? В [отчете OWASP Top Ten](https://owasp.org/www-project-top-ten/) взлом аутентификации 
находится на **втором месте**!

> 2. Взлом аутентификации. Функции приложений, связанные с аутентификацией и 
> управлением сеансами часто реализуются неправильно, что позволяет злоумышленникам 
> скомпрометировать пароли, ключи или токены сеанса, или использовать другие 
> недостатки реализации для временного или постоянного доступа к учётным записям
> других пользователей.

Вы можете сказать: _"Моя компания просто продает тостеры! Почему мне должно быть 
до этого дело?"_. Я почти уверен, что вам по-прежнему небезразличен имидж своей 
компании. **Дискредитация после инцидента, вызванного хакерским взломом 
аккаунтов ваших клиентов, всегда доставляет неудобства.**

Вам всё ещё кажется, что Вы можете реализовать идеальную аутентификацию?
Даже гиганты, в которых сотни исследователей и разработчиков работают только 
над аутентификацией, не могут этого сделать. В марте 2020 года исследователь 
[обнаружил уязвимость Facebook](https://latesthackingnews.com/2020/03/03/10-year-old-facebook-oauth-framework-flaw-discovered/), которая может позволить взломать чью-либо 
учетную запись Facebook. Всего несколько месяцев спустя ошибка в механизме 
входа в Apple [позволила полностью завладеть учетными записями пользователей этого 
стороннего приложения](https://bhavukjain.com/blog/2020/05/30/zeroday-signin-with-apple/).

Если вы всё ещё не уверены, стоит подумать об экономии вашего времени. Я уверен, 
что **клиенты вашей компании будут намного счастливее, если подарите им долгожданный
функционал. А не навороченную, кастомную аутентификацию 😉.**

## Использование Firebase аутентификации

Конечно, существует множество решений, которые можно использовать для 
реализации аутентификации. В этой статье мы хотим предложить самый быстрый и 
простой способ для нашего случая. Мы также не рассматриваем какие-либо проблемы, 
связанные с привязкой к поставщику услуг — мы займемся этим в будущем 😉.

> Если вы ищете простое решение, не зависящее от какого-либо провайдера, вам
> следует обратить внимание на [JWT](https://jwt.io/) и [библиотеку dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go).

> Это не просто очередная статья со случайными фрагментами кода.
> 
> Этот пост является частью большого цикла, показывающий как создавать приложения на
> **Go, которые легко разрабатывать, поддерживать и с ними интересно работать в
> долгосрочной перспективе**. Мы делаем это, делясь проверенными методами, основанными
> на многих экспериментах, проведёнными с возглавляемыми нами с командами,
> и [научных исследованиях](https://threedots.tech/post/ddd-lite-in-go-introduction/#thats-great-but-do-you-have-any-evidence-it-works).
> 
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
>
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

Приложение-пример Wild Workouts, которое мы создали для этой серии статей, в настоящее
время размещено на хостинге Firebase. Использование [Firebase аутентификации](https://firebase.google.com/docs/auth) - 
естественный выбор. У этого решения есть одно существенное преимущество — оно 
работает почти полностью из коробки как со стороны бэкэнда, так и со стороны 
фронтенда.

> Развертывание проекта подробно описано в статье _Подробное описание 
> настройки Terraform для бессерверного приложения в Google Cloud Run и Firebase_.
> В [первой статье](https://threedots.tech/post/serverless-cloud-run-firebase-modern-go-application/) в конце есть раздел, касающийся развертывания.
> Во время настройки не забудьте включить провайдер для входа по электронной 
> почте/паролю на вкладке **Аутентификация/Метод входа в систему** в [консоли 
> Firebase](https://console.firebase.google.com/project/_/authentication/providers)!
> 
> ![firebase-console](images/part4/firebase-auth-providers.png)

## Фронтенд

Первое, что нам нужно сделать, это [инициализировать Firebase SDK](https://firebase.google.com/docs/web/setup#add-sdks-initialize).

Далее в форме на главной странице вызываем функцию `loginUser`. Эта функция 
вызывает `Auth.login`, `Auth.waitForAuthReady` и `Auth.getJwtToken`. Результат
присваивается клиентам, сгенерированным OpenAPI, с помощью `setApiClientsAuth`.

```js
// ...
export function loginUser(login, password) {
    return Auth.login(login, password)
        .then(function () {
            return Auth.waitForAuthReady()
        })
        .then(function () {
            return Auth.getJwtToken(false)
        })
        .then(token => {
            setApiClientsAuth(token)
        })
// ...
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/web/src/repositories/user.js](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/4de7a36ee40e4cff5334bc33597592cc8468df54/web/src/repositories/user.js#L36)

`Auth` - это класс с двумя реализациями: Firebase и mock. Позже мы рассмотрим 
фиктивную реализацию, а сейчас сосредоточимся на Firebase.

Чтобы войти в систему нам нужно вызвать `firebase.auth().SignInWithEmailAndPassword`.
Если все в порядке, `auth().CurrentUser.getIdToken` вернёт наш JWT токен.

```js
class FirebaseAuth {
    login(login, password) {
        return firebase.auth().signInWithEmailAndPassword(login, password)
    }

    waitForAuthReady() {
        return new Promise((resolve) => {
            firebase
                .auth()
                .onAuthStateChanged(function () {
                    resolve()
                });
        })
    }

    getJwtToken(required) {
        return new Promise((resolve, reject) => {
            if (!firebase.auth().currentUser) {
                if (required) {
                    reject('no user found')
                } else {
                    resolve(null)
                }
                return
            }

            firebase.auth().currentUser.getIdToken(false)
                .then(function (idToken) {
                    resolve(idToken)
                })
                .catch(function (error) {
                    reject(error)
                });
        })
    }

    // ...
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/web/src/repositories/auth.js](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/4de7a36ee40e4cff5334bc33597592cc8468df54/web/src/repositories/auth.js#L6)

> Firebase также поддерживает из коробки вход в систему с помощью самых популярных 
> OAuth провайдеров, таких как Facebook, Gmail или GitHub.

Затем нам нужно присвоить атрибуту `authentications['bearerAuth'].accessToken`
клиентам, [сгенерированным OpenAPI](https://threedots.tech/post/serverless-cloud-run-firebase-modern-go-application/#public-http-api), JWT токен, полученный из 
`Auth.getJwtToken(false)`. Присваиваем этот токен OpenAPI клиентам и вуаля! Все 
наши запросы теперь аутентифицированы.

```js
export function setApiClientsAuth(idToken) {
    usersClient.authentications['bearerAuth'].accessToken = idToken
    trainerClient.authentications['bearerAuth'].accessToken = idToken
    trainingsClient.authentications['bearerAuth'].accessToken = idToken
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/web/src/repositories/auth.js](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/4de7a36ee40e4cff5334bc33597592cc8468df54/web/src/repositories/auth.js#L118)

> Если Вы создаёте свою собственную спецификацию OpenAPI, то она не будет 
> работать без соответствующего [определения параметра аутентификации](https://swagger.io/docs/specification/authentication/bearer-authentication/). В 
> [спецификации OpenAPI](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/a0a41253db96d46d75e7ff4c7e7f95848f47dcc3/api/openapi/trainings.yml#L14) Wild Workouts это уже сделано.

Если вы хотите узнать больше, я бы порекомендовал вам просмотреть [ссылку по API 
Firebase аутентификации](https://firebase.google.com/docs/reference/js/firebase.auth).

## Бекенд

Бекенд фактически использует эту аутентификацию на нашем HTTP-сервере. Я создал 
простой HTTP middleware, который сделает эту задачу за нас.

Этот middleware выполняет три функции:

1. Получает токен из HTTP заголовка
2. Проверяет токен с помощью клиента Firebase аутентификации
3. Сохраняет данные пользователя в контексте

```go
package auth

import (
    "context"
    "net/http"
    
    "firebase.google.com/go/auth"
    "github.com/MaksimDzhangirov/three-dots/internal/common/server/httperr"
)

type FirebaseHttpMiddleware struct {
    AuthClient *auth.Client
}

func (a FirebaseHttpMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
    
        bearerToken := a.tokenFromHeader(r)
        if bearerToken == "" {
            httperr.Unauthorised("empty-bearer-token", nil, w, r)
            return
        }
    
        token, err := a.AuthClient.VerifyIDToken(ctx, bearerToken)
        if err != nil {
            httperr.Unauthorised("unable-to-verity-jwt", err, w, r)
            return
        }
    
        // всегда рекомендуется использовать пользовательский тип в качестве значения контекста (в данном случае ctxKey)
        // потому что никто из вне пакета не сможет переопределить/прочитать это значение
        ctx = context.WithValue(ctx, userContentKey, User{
            UUID:        token.UID,
            Email:       token.Claims["email"].(string),
            Role:        token.Claims["role"].(string),
            DisplayName: token.Claims["name"].(string),
        })
        r = r.WithContext(ctx)
    
        next.ServeHTTP(w, r)
    })
}

// ...

type ctxKey int

const (
    userContentKey ctxKey = iota
)

// ...

func UserFromCtx(ctx context.Context) (User, error) {
    u, ok := ctx.Value(userContentKey).(User)
    if ok {
        return u, nil
    }
    
    return User{}, NoUserInContextError
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/common/auth/http.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/33be9ae12a031b85b09fd8779fc445445785b106/internal/common/auth/http.go#L13)

Данные пользователя теперь будут в каждом HTTP-запросе с помощью функции 
`auth.UserFromCtx`.

```go
func (h HttpServer) GetTrainings(w http.ResponseWriter, r *http.Request) {
    user, err := auth.UserFromCtx(r.Context())
    if err != nil {
        httperr.Unauthorised("no-user-found", err, w, r)
        return
    }

    // ...
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainings/http.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/33be9ae12a031b85b09fd8779fc445445785b106/internal/trainings/http.go#L26)

Мы также можем ограничить доступ к некоторым ресурсам, в зависимости от прав
пользователя.

```go
func (h HttpServer) MakeHourAvailable(w http.ResponseWriter, r *http.Request) {
    user, err := auth.UserFromCtx(r.Context())
    if err != nil {
        httperr.Unauthorised("no-user-found", err, w, r)
        return
    }

    if user.Role != "trainer" {
        httperr.Unauthorised("invalid-role", nil, w, r)
        return
    }

    // ...
}
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/http.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/ef6056fdeb39b89009127e07600f1b9ec87e717c/internal/trainer/http.go#L32)

## Добавление пользователей

В нашем случае мы добавляем пользователей во время запуска сервиса `users`. Вы 
можете также добавить их из FirebaseUI. К сожалению, через UI вы не можете 
установить все необходимые данные, например claims - это нужно делать через API.

```go
config := &firebase.Config{ProjectID: os.Getenv("GCP_PROJECT")}
firebaseApp, err := firebase.NewApp(context.Background(), config, opts...)
if err != nil {
    return nil, err
}

authClient, err := firebaseApp.Auth(context.Background())
if err != nil {
    return nil, err
}

// ...

for _, user := range usersToCreate {
    userToCreate := (&auth.UserToCreate{}).
        Email(user.Email).
        Password("123456").
        DisplayName(user.DisplayName)

    createdUser, err := authClient.CreateUser(context.Background(), userToCreate)
    
    // ...

    err = authClient.SetCustomUserClaims(context.Background(), createdUser.UID, map[string]interface{}{
        "role": user.Role,
    })
    
    // ...
}
// ...
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/users/fixtures.go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/ef6056fdeb39b89009127e07600f1b9ec87e717c/internal/users/fixtures.go#L103)

![firebase-auth](images/part4/firebase-auth.png)

## Имитация аутентификации для локальной разработки

Существует высокая потребность и [обсуждается проблема](https://github.com/firebase/firebase-tools/issues/1677) поддержки эмулятора
Firebase аутентификации. К сожалению, его пока не существует. Здесь ситуация 
очень похожа на Firestore - я хочу иметь возможность запускать своё приложение 
локально без каких-либо внешних зависимостей. Пока не существует эмулятора, нет
другого способа кроме как реализовать простую фиктивную реализацию.

В реализации для бекенда и фронтенда нет ничего сложного. GetJwtToken [на 
фронтенде](../code/wild-workouts/part4/web/src/repositories/auth.js) реализуется 
путём простой генерации JWT токена с `mock_secret`. [Бекенд](../code/wild-workouts/part4/internal/common/auth/http_mock.go) вместо вызова Firebase
для проверки токена, проверяет, был ли сгенерирован JWT с `mock_secret`.

Это вселяет уверенность, что наш алгоритм реализован более или менее правильно. 
Но можно ли протестировать его с помощью на Firebase продакшен локально?

## Firebase аутентификация для локальной разработки

Фиктивная аутентификация не даёт нам 100% уверенности в том, что алгоритм работает
правильно. У нас должна быть возможность протестировать Firebase аутентификацию локально.
Для этого вам нужно проделать несколько дополнительных шагов.

**Это не так просто. Если вы ничего не меняете в аутентификации, вы, вероятно, 
можете [пропустить эту часть](part4.md#firebase----).**

Прежде всего, вам необходимо сгенерировать файл служебной учетной записи в репозитории.

> Помните, что не нужно добавлять его в репозиторий! Не волнуйтесь, он уже в `.gitignore`.

Далее нужно раскомментировать **все** строки, добавляющие служебную учетную запись
в docker-compose.yml.

```yaml
     volumes:
       - ./pkg:/pkg
-#      - ./service-account-file.json:$SERVICE_ACCOUNT_FILE
+      - ./service-account-file.json:$SERVICE_ACCOUNT_FILE
     working_dir: /pkg/trainer
     ports:
       - "127.0.0.1:3000:$PORT"

       context: ./dev/docker/app
     volumes:
       - ./pkg:/pkg
-#      - ./service-account-file.json:$SERVICE_ACCOUNT_FILE
+      - ./service-account-file.json:$SERVICE_ACCOUNT_FILE
     working_dir: /pkg/trainer

@ ... сделайте это для всех сервисов!
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/docker-compose.yml](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/3d2ce607c0bcddc2a7f9cb56193167c8543d57ec/docker-compose.yml)

После этого вы должны присвоить в `GCP_PROJECT` значению своего идентификатора 
проекта, раскомментировать `SERIVCE_ACCOUNT_FILE` и установить `MOCK_AUTH` в 
`false` в `.env`.

```
-GCP_PROJECT=threedotslabs-cloudnative
+GCP_PROJECT=YOUR-PROJECT-ID
 
 PORT=3000
 
-FIRESTORE_PROJECT_ID=threedotslabs-cloudnative
+FIRESTORE_PROJECT_ID=YOUR-PROJECT-ID
 FIRESTORE_EMULATOR_HOST=firestore:8787
 # env used by karhoo/firestore-emulator container
-GCP_PROJECT_ID=threedotslabs-cloudnative
+GCP_PROJECT_ID=YOUR-PROJECT-ID
 
 TRAINER_GRPC_ADDR=trainer-grpc:3000
 USERS_GRPC_ADDR=users-grpc:3000

@ ...
 
 CORS_ALLOWED_ORIGINS=http://localhost:8080
 
-#SERVICE_ACCOUNT_FILE=/service-account-file.json
-MOCK_AUTH=true
+SERVICE_ACCOUNT_FILE=/service-account-file.json
+MOCK_AUTH=false
 LOCAL_ENV=true
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/.env](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/3d2ce607c0bcddc2a7f9cb56193167c8543d57ec/.env)

```js
-const MOCK_AUTH = process.env.NODE_ENV === 'development'
+const MOCK_AUTH = false
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/web/src/repositories/auth.js](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/4de7a36ee40e4cff5334bc33597592cc8468df54/web/src/repositories/auth.js#L124)

Теперь вам нужно перейти в [Настройки Проекта](https://console.firebase.google.com/project/_/settings/general) (для этого щелкните на
значок шестеренки справа от _Project overview_) и обновить настройки Firebase в 
`web/public/__/firebase/init.json`.

![settings](images/part4/firebase-app-config.png)

```json
-  "apiKey": "",
-  "appId": "",
-  "authDomain": "",
-  "databaseURL": "",
-  "messagingSenderId": "",
-  "projectId": "",
-  "storageBucket": ""
+  "apiKey": "AIzaSyBW6TNG5Xdq1F7MleYml8CovOej5SbxXhw",
+  "authDomain": "wild-workouts-6.firebaseapp.com",
+  "databaseURL": "https://wild-workouts-6.firebaseio.com",
+  "projectId": "wild-workouts-6",
+  "storageBucket": "wild-workouts-6.appspot.com",
+  "messagingSenderId": "519519305315",
+  "appId": "1:519519305315:web:cde76e4847f0d95cc84256"
 }
```
Весь исходный код: [github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/web/public/__/firebase/init.json](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/4de7a36ee40e4cff5334bc33597592cc8468df54/web/public/__/firebase/init.json)

**Последнее, что нужно сделать, это остановить и снова запустить docker-compose** 
для перезагрузки переменных окружения и ссылок.

## И на этом всё

Как видите, настройка была действительно простой. Нам удалось сэкономить много 
времени. **Ситуация могла бы быть более сложной, если бы мы не хотели размещать 
наше приложение на Firebase.** Наша идея с самого начала заключалась в том, 
чтобы подготовить простую основу для будущих статей.

В этой статье мы не исчерпали тему аутентификации. Но, по крайней мере, на 
данный момент мы не планируем углубляться в аутентификацию.

**Может быть, у вас есть хороший или плохой опыт использования любого решения 
для аутентификации в Go? Пожалуйста, поделитесь в комментариях, чтобы все 
читатели могли извлечь из этого пользу!** 👍

В этой серии статей мы хотели бы показать вам, как создавать приложения, которые
легко разрабатывать, поддерживать и приятно работать в долгосрочной перспективе.
Возможно, вы видели примеры доменно-ориентированного проектирования или чистой 
архитектуры в Go. Большинство из них реализованы не так прагматично, как это 
выглядит на словах.

**В последующих статьях мы покажем методики, успешно используемые в командах, 
которыми руководим в течение нескольких лет.** Мы также рассмотрим случаи, когда их 
применение имеет смысл, а когда это излишество или разработка ради строчки в 
резюме. Я уверен, что это заставит вас пересмотреть свою точку зрения на эти 
методики. 😉

Это все на сегодня. Следующая статья выйдет через 1-2 недели. 
**Подпишитесь на нашу рассылку, если вы не хотите её пропустить!** 😉