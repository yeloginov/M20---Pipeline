FROM golang
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
# Добавляем файлы main.go и go.mod в папку /go/src/pipeline
ADD main.go .
ADD go.mod .
# Компилируем программы и помещаем запускаемый файл в /go/bin/
RUN go install .