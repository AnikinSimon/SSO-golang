# SSO-golang

В рамках проекта SSO-golang реализован сервис по аутенфикации и авторизации.
Взаимодействовать можно при помощи gRPC сервера и gRPC-шлюза, которые предоставляют следующие возможности:
* Регистрации пользователей в сторонных приложениях
* Операция Login и выдача JWT-токена
* Проверка на то, является ли пользователь администратором

Соединение с сервером защищено взаимным TLS шифрованием. 
В ветке ```feature/server-side-tls``` реализовано шифрование со стороны сервера

# Описание интерфейса

Описание содержиться в папке ```api/```

1. RegisterApp
    * Регистрация приложения для последующего добавления пользователей
    * Запрос RegisterAppRequest 
        * string name = 1;
        * string secret = 2;
    * Ответ RegisterAppResponse 
        * string app_uuid = 1; 

2. RegisterApp
    * Регистрация пользователя 
    * Запрос RegisterRequest 
        * string email = 1;
        * string password = 2;
        * string app_uuid = 3;
    * Ответ RegisterResponse 
        * string user_uuid = 1; 

3. Login
    * Аутенфикация пользователя 
    * Запрос LoginRequest 
        * string email = 1; 
        * string password = 2;
        * string app_uuid = 3; 
    * Ответ LoginResponse 
        * string token = 1; 

4. IsAdmin
    * Проверка является ли пользователь администратором
    * Запрос IsAdminRequest 
        * string user_uuid = 1; 
    * Ответ IsAdminResponse 
        * bool is_admin = 1; 

# Технологический стек
Golang, Postgres, gRPC, GORM, Protobuf, JWT, gRPC-Gateway

# Генерация сертификатов
```
make cert
```

# Запуск
```
make run
```

# Остановка

```
make stop
```

