FROM golang

WORKDIR /go/src

RUN go get -u github.com/gorilla/mux
RUN go get -u github.com/gorilla/securecookie
RUN go get -u github.com/gorilla/sessions
RUN go get -u github.com/go-sql-driver/mysql
RUN go get -u github.com/joho/godotenv
RUN go get -u golang.org/x/crypto/bcrypt

COPY . .

CMD ["go", "run", "."]