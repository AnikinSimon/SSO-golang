rm *.pem

# 1. Генерируем приватный ключ CA и самоподписанный сертификат
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem \
-subj "/C=RU/ST=Moscow/emailAddress=test.email@gmail.com"

# 2. Генерируем приватный ключ веб-сервера и запрос на подпись сертификата (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem \
-subj "/C=RU/ST=Moscow/emailAddress=test.server@gmail.com"

# 3. Используем приватный ключ CA, чтобы подписать CSR веб-сервера и получить обратно подписанный сертификат 
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem \
-CAcreateserial -out server-cert.pem -extfile server-ext.cnf
